package gofis

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
)

var StaticArr = make(map[string][]string)
var AsyncArr = make(map[string]interface{})
var Framework string
var loaded = make(map[string]string)
var arrMap = make(map[string]interface{})

func register(ns, root string) bool {
	mapjson := root + ns + "-map.json"
	content, err := ioutil.ReadFile(mapjson)

	if err != nil {
		log.Println("Can't found: "+mapjson, err)
		return false
	}

	buffer := bytes.NewBuffer(content)

	dec := json.NewDecoder(buffer)

	var res map[string]interface{}

	err = dec.Decode(&res)

	if err != nil {
		log.Println(err)
		return false
	}

	arrMap[ns] = res

	return true
}

func delAsyncDeps(id string) {
	arrRes := AsyncArr[id].(map[string]interface{})
	if pkg, ok := arrRes["pkg"]; ok {
		arrPkg, ok := AsyncArr[pkg.(string)]
		if ok {
			StaticArr["js"] = append(StaticArr["js"], arrPkg.(map[string]interface{})["uri"].(string))
			delete(AsyncArr, pkg.(string))
			arrHas := arrPkg.(map[string]interface{})["has"]
			for _, hasId := range arrHas.([]interface{}) {
				delAsyncDeps(hasId.(string))
			}
		} else {
			arrHas := arrPkg.(map[string]interface{})["has"]
			for _, hasId := range arrHas.([]interface{}) {
				delete(AsyncArr, hasId.(string))
			}
		}
	} else {
		StaticArr["js"] = append(StaticArr["js"], arrRes["uri"].(string))
		delete(AsyncArr, id)
	}

	if arrDeps, _ok := arrRes["deps"]; _ok {
		for _, depId := range arrDeps.([]interface{}) {
			delAsyncDeps(depId.(string))
		}
	}
}

func loadDeps(arrRes map[string]interface{}, root string, async bool) {
	var deps interface{}
	var asyncs interface{}
	var extras interface{}
	var ok bool

	if extras, ok = arrRes["extras"]; ok {
		if asyncs, ok = extras.(map[string]interface{})["async"]; ok {
			for _, asyncId := range asyncs.([]interface{}) {
				Load(asyncId.(string), root, async)
			}
		}
	}

	if deps, ok = arrRes["deps"]; ok {
		for _, id := range deps.([]interface{}) {
			Load(id.(string), root, async)
		}
	}
}

func GetUri(id, root string) string {
	if uri, ok := loaded[id]; ok {
		return uri
	} else {
		var ns string
		if p := strings.Index(id, ":"); p != -1 {
			ns = id[0:p]
		}

		if _, ok := arrMap[ns]; ok || register(ns, root) {
			resMap := arrMap[ns].(map[string]interface{})
			arrRes, ok := resMap["res"].(map[string]interface{})[id]
			if ok {
				res := arrRes.(map[string]interface{})
				if pkg, ok := res["pkg"]; ok {
					arrPkg := resMap[pkg.(string)].(map[string]interface{})
					return arrPkg["uri"].(string)
				} else {
					return res["uri"].(string)
				}
			}
		}
	}
	log.Println("Can't found id: " + id)
	return ""
}

/**
* load resource from id and load resource deps
 */
func Load(id, root string, async bool) string {
	var ok bool
	var uri string
	if uri, ok = loaded[id]; ok {
		return uri
	} else {
		var ns string
		if p := strings.Index(id, ":"); p != -1 {
			ns = id[0:p]
		} else {
			log.Println("namespace not exists. >id: " + id)
		}

		if _, ok = arrMap[ns]; ok || register(ns, root) {
			resMap := arrMap[ns].(map[string]interface{})
			if _, ok = resMap["res"]; ok {
				if arrRes, resExists := resMap["res"].(map[string]interface{})[id]; resExists {
					if pkg, pkgExists := arrRes.(map[string]interface{})["pkg"]; pkgExists {
						arrPkg := resMap["pkg"].(map[string]interface{})[pkg.(string)]
						uri = arrPkg.(map[string]string)["uri"]
					} else {
						uri = arrRes.(map[string]interface{})["uri"].(string)
						loaded[id] = uri
						loadDeps(arrRes.(map[string]interface{}), root, async)
					}

					if async {

					} else {
						_type := arrRes.(map[string]interface{})["type"].(string)
						StaticArr[_type] = append(StaticArr[_type], uri)
					}
					return uri
				}
			} else {
				log.Println("Can't found resource")
				return ""
			}
		}
	}
	return ""
}
