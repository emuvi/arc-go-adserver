package guide

import (
	"fmt"
	"strconv"
	"sync"
)

type aConfigs struct {
	data  map[string]string
	mutex sync.Mutex
}

var Configs = &aConfigs{
	data:  map[string]string{},
	mutex: sync.Mutex{},
}

func (configs *aConfigs) GetString(key string, standard string) string {
	configs.mutex.Lock()
	defer configs.mutex.Unlock()
	result, ok := configs.data[key]
	if ok {
		return result
	}
	return standard
}

func (configs *aConfigs) SetString(key string, value string) {
	configs.mutex.Lock()
	defer configs.mutex.Unlock()
	configs.data[key] = value
}

func (configs *aConfigs) GetBool(key string, standard bool) bool {
	configs.mutex.Lock()
	defer configs.mutex.Unlock()
	result, ok := configs.data[key]
	if ok {
		converted, err := strconv.ParseBool(result)
		if err == nil {
			return converted
		}
	}
	return standard
}

func (configs *aConfigs) SetBool(key string, value bool) {
	configs.mutex.Lock()
	defer configs.mutex.Unlock()
	configs.data[key] = strconv.FormatBool(value)
}

func (configs *aConfigs) GetInt(key string, standard int) int {
	configs.mutex.Lock()
	defer configs.mutex.Unlock()
	result, ok := configs.data[key]
	if ok {
		converted, err := strconv.Atoi(result)
		if err == nil {
			return converted
		}
	}
	return standard
}

func (configs *aConfigs) SetInt(key string, value int) {
	configs.mutex.Lock()
	defer configs.mutex.Unlock()
	configs.data[key] = strconv.Itoa(value)
}

func (configs *aConfigs) GetFloat(key string, standard float64) float64 {
	configs.mutex.Lock()
	defer configs.mutex.Unlock()
	result, ok := configs.data[key]
	if ok {
		converted, err := strconv.ParseFloat(result, 64)
		if err == nil {
			return converted
		}
	}
	return standard
}

func (configs *aConfigs) SetFloat(key string, value float64) {
	configs.mutex.Lock()
	defer configs.mutex.Unlock()
	configs.data[key] = fmt.Sprintf("%g", value)
}
