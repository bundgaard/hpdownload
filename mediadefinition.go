package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/bundgaard/js/object"
)

func walkHashMap(anHashMap *object.Hash) {
	for _, value := range anHashMap.Pairs {
		realValue := value.Key.(*object.StringObject)
		if realValue.Value == "mediaDefinitions" {
			fmt.Println(value.Key.Inspect(), value.Value.Inspect())
			walkMediaDefinitions(value.Value.(*object.Array))
		}

	}
}

func walkMediaDefinitions(anArray *object.Array) {
	for _, val := range anArray.Elements {
		switch v := val.(type) {
		case *object.Hash:
			walkMediaDefinitionPairs(v)
		}
	}

}

type mediaDefinition struct {
	DefaultQuality bool
	VideoURL       string
	Quality        int
	Qualities      []int
	Format         string
}

var mediadefinitions = make([]mediaDefinition, 0)

func getKeyValue(str *object.StringObject) string {
	return str.Value
}
func getBoolValue(str *object.Boolean) bool {
	return str.Value
}
func walkArray(arr *object.Array) []string {
	output := make([]string, 0)
	for _, x := range arr.Elements {
		switch xType := x.(type) {
		case *object.StringObject:
			output = append(output, xType.Value)
		case *object.NumberObject:

		default:
			log.Fatalf("walkArray not handled: %v %T", x.Inspect(), xType)
		}
	}
	return output
}
func walkMediaDefinitionPairs(v *object.Hash) {
	var md mediaDefinition
	for _, v := range v.Pairs {

		keyObj, ok := v.Key.(*object.StringObject)
		if !ok {
			log.Fatal("failed to convert key")
		}
		key := getKeyValue(keyObj)
		if key == "defaultQuality" {
			boolObj, ok := v.Value.(*object.Boolean)
			if ok {
				md.DefaultQuality = getBoolValue(boolObj)
			}
			numberObj, ok := v.Value.(*object.NumberObject)
			if ok {
				md.Quality = int(numberObj.Value)
			}

		}
		if key == "format" {
			md.Format = getKeyValue(v.Value.(*object.StringObject))
		}
		if key == "videoUrl" {
			md.VideoURL = getKeyValue(v.Value.(*object.StringObject))
		}
		if key == "quality" {
			switch xValue := v.Value.(type) {
			case *object.StringObject:
				qualityN, _ := strconv.Atoi(xValue.Value)
				md.Quality = qualityN
			case *object.Array:
				xs := walkArray(xValue)
				for _, x := range xs {
					tmpN, _ := strconv.Atoi(x)
					md.Qualities = append(md.Qualities, tmpN)
				}
			}
		}

	}
	mediadefinitions = append(mediadefinitions, md)
}
