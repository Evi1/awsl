package manage

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Evi1/awsl/config"
	"github.com/Evi1/awsl/tools"
)

// ServerFlowManager ServerFlowManager
var ServerFlowManager *SFlowManager

var limit int

func init() {
	limit = int(30 * 24 * time.Hour / time.Second)
	ServerFlowManager = &SFlowManager{in: make(map[int]map[string]*tools.Counter),
		out:        make(map[int]map[string]*tools.Counter),
		inHistory:  make(map[int]map[string][]uint64, 0),
		outHistory: make(map[int]map[string][]uint64, 0),
		lock:       sync.RWMutex{}}
	config.GetConf()
	if config.Manage > 0 {
		go ServerFlowManager.Tick()
	}
}

// SFlowManager SFlowManager
type SFlowManager struct {
	in         map[int]map[string]*tools.Counter
	out        map[int]map[string]*tools.Counter
	inHistory  map[int]map[string][]uint64
	outHistory map[int]map[string][]uint64
	inSum      uint64
	outSum     uint64
	lock       sync.RWMutex
}

func (fm *SFlowManager) add(id int, host string, count int64, m map[int]map[string]*tools.Counter, history map[int]map[string][]uint64) {
	hostMap, ok := m[id]
	if !ok {
		fm.lock.Lock()
		hostMap, ok = m[id]
		if !ok {
			m[id] = make(map[string]*tools.Counter)
			history[id] = make(map[string][]uint64)
			hostMap = m[id]
		}
		fm.lock.Unlock()
	}
	counter, ok := hostMap[host]
	if !ok {
		fm.lock.Lock()
		counter, ok = hostMap[host]
		if !ok {
			hostMap[host] = tools.NewCounter()
			history[id][host] = make([]uint64, 0)
			counter = hostMap[host]
		}
		fm.lock.Unlock()
	}
	counter.Add(count)
}

// AddIn AddIn
func (fm *SFlowManager) AddIn(id int, host string, count int64) {
	fm.add(id, host, count, fm.in, fm.inHistory)
	fm.inSum += uint64(count)
}

// AddOut AddOut
func (fm *SFlowManager) AddOut(id int, host string, count int64) {
	fm.add(id, host, count, fm.out, fm.outHistory)
	fm.outSum += uint64(count)
}

func (fm *SFlowManager) tickFor(m map[int]map[string]*tools.Counter, history map[int]map[string][]uint64) {
	for id := range m {
		for host := range m[id] {
			fm.lock.RLock()
			counter := m[id][host]
			fm.lock.RUnlock()
			num := counter.GetSet(0)
			fm.lock.Lock()
			history[id][host] = append([]uint64{uint64(num)}, history[id][host]...)
			if len(history[id][host]) > limit {
				history[id][host] = history[id][host][:limit]
			}
			fm.lock.Unlock()
		}
	}
}

// Tick Tick
func (fm *SFlowManager) Tick() {
	t := time.Tick(time.Second)
	for {
		select {
		case <-t:
			go fm.tickFor(fm.in, fm.inHistory)
			go fm.tickFor(fm.out, fm.outHistory)
		}
	}
}

// GetRoot GetRoot
func (fm *SFlowManager) GetRoot() string {
	m := make(map[string]string)
	m["in sum"] = handleBytesNum(fm.inSum)
	m["out sum"] = handleBytesNum(fm.outSum)
	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

type sIDMap struct {
	In      string
	Out     string
	SumIns  string
	SumOuts string
}

// GetID GetID
func (fm *SFlowManager) GetID(id int) string {
	resultMap := make(map[string]sIDMap)
	fm.lock.RLock()
	inMap := fm.inHistory[id]
	outMap := fm.outHistory[id]
	sumIn := uint64(0)
	sumOut := uint64(0)
	sumIns := uint64(0)
	sumOuts := uint64(0)
	for k, inv := range inMap {
		outv := outMap[k]
		if len(inv) == 0 || len(outv) == 0 {
			continue
		}
		sumInForK := uint64(0)
		sumOutForK := uint64(0)
		for _, v := range inv {
			sumInForK += v
		}
		for _, v := range outv {
			sumOutForK += v
		}
		resultMap[k] = sIDMap{In: handleBytesNum(inv[0]), Out: handleBytesNum(outv[0]), SumIns: handleBytesNum(sumInForK), SumOuts: handleBytesNum(sumOutForK)}
		sumIn += inv[0]
		sumOut += outv[0]
		sumIns += sumInForK
		sumOuts += sumOutForK
	}
	fm.lock.RUnlock()
	resultMap["0"] = sIDMap{In: handleBytesNum(sumIn), Out: handleBytesNum(sumOut), SumIns: handleBytesNum(sumIns), SumOuts: handleBytesNum(sumOuts)}
	bytes, err := json.MarshalIndent(resultMap, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

// GetIDHistory GetIDHistory
func (fm *SFlowManager) GetIDHistory(id int) string {
	fm.lock.RLock()
	defer fm.lock.RUnlock()
	inMap := fm.inHistory[id]
	outMap := fm.outHistory[id]
	res := "{\n"
	for k, inv := range inMap {
		outv := outMap[k]
		resM := make(map[string][]uint64)
		resM["ins"] = make([]uint64, len(inv))
		resM["outs"] = make([]uint64, len(outv))
		copy(resM["ins"], inv)
		copy(resM["outs"], outv)
		bytes, err := json.Marshal(resM)
		if err != nil {
			res += "\"" + k + "\":" + err.Error() + ",\n"
			continue
		}
		res += "\"" + k + "\":" + string(bytes) + ",\n"
	}
	li := strings.LastIndex(res, ",\n")
	if li >= 0 {
		res = res[:li] + res[li+1:]
	}
	return res + "}"
}

var mark = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

func handleBytesNum(num uint64) string {
	i := 0
	fNum := float64(num)
	for i < len(mark)-1 && fNum > 1024 {
		i++
		fNum /= 1024
	}
	str := strconv.FormatFloat(fNum, 'f', 2, 64)
	if strings.HasSuffix(str, ".00") {
		str = str[:len(str)-3]
	}
	return str + mark[i]
}

func handleBytesNumList(list []uint64) []string {
	res := make([]string, len(list))
	for i, v := range list {
		res[i] = handleBytesNum(v)
	}
	return res
}
