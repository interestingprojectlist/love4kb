package mapreduce

import (
	"encoding/json"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
)

func doMap(
	jobName string, // the name of the MapReduce job
	mapTask int, // which map task this is
	inFile string,
	nReduce int, // the number of reduce task that will be run ("R" in the paper)
	mapF func(filename string, contents string) []KeyValue,
) {
	//
	// doMap manages one map task: it should read one of the input files
	// (inFile), call the user-defined map function (mapF) for that file's
	// contents, and partition mapF's output into nReduce intermediate files.
	//
	// There is one intermediate file per reduce task. The file name
	// includes both the map task number and the reduce task number. Use
	// the filename generated by reduceName(jobName, mapTask, r)
	// as the intermediate file for reduce task r. Call ihash() (see
	// below) on each key, mod nReduce, to pick r for a key/value pair.
	//
	// mapF() is the map function provided by the application. The first
	// argument should be the input file name, though the map function
	// typically ignores it. The second argument should be the entire
	// input file contents. mapF() returns a slice containing the
	// key/value pairs for reduce; see common.go for the definition of
	// KeyValue.
	//
	// Look at Go's ioutil and os packages for functions to read
	// and write files.
	//
	// Coming up with a scheme for how to format the key/value pairs on
	// disk can be tricky, especially when taking into account that both
	// keys and values could contain newlines, quotes, and any other
	// character you can think of.
	//
	// One format often used for serializing data to a byte stream that the
	// other end can correctly reconstruct is JSON. You are not required to
	// use JSON, but as the output of the reduce tasks *must* be JSON,
	// familiarizing yourself with it here may prove useful. You can write
	// out a data structure as a JSON string to a file using the commented
	// code below. The corresponding decoding functions can be found in
	// common_reduce.go.
	//
	//   enc := json.NewEncoder(file)
	//   for _, kv := ... {
	//     err := enc.Encode(&kv)
	//
	// Remember to close the file after you have written all the values!
	//
	// Your code here (Part I).
	//

	inputFile, err := os.Open(inFile)
	if err != nil {
		log.Panicf("open file:%s cause error:%s\n", inFile, err.Error())
		return
	}
	defer func() {
		if err := inputFile.Close(); err != nil {
			log.Panicf("close file:%s cause error:%s\n", inFile, err.Error())
		}
	}()

	inputContent, err := ioutil.ReadAll(inputFile)
	if err != nil  {
		log.Panicf("read file:%s cause error:%s\n", inFile, err.Error())
		return
	}

	fileIndies := make(map[string][]int)
	kvList := mapF(inFile, string(inputContent))
	for index, kvPair := range kvList {
		r := ihash(kvPair.Key) % nReduce
		reduceFileName := reduceName(jobName, mapTask, r)

		if indexList, ok := fileIndies[reduceFileName]; ok {
			fileIndies[reduceFileName] = append(indexList, index)
		} else {
			fileIndies[reduceFileName] = []int{index}
		}
	}

	for fileName, indies := range fileIndies {
		func() {
			outFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				log.Panicf("open file:%s cause error:%s\n", fileName, err.Error())
				return
			}
			func() {
				defer func() {
					if err := outFile.Close(); err != nil {
						log.Panicf("close file:%s cause error:%s\n",fileName, err.Error())
					}
				}()

				enc := json.NewEncoder(outFile)
				for _, index := range indies {
					kvPair := kvList[index]
					if err := enc.Encode(KeyValue{kvPair.Key, kvPair.Value}); err != nil {
						log.Panicf("encode error:%s\n", err.Error())
					}
				}
			}()
		}()
	}
}

func ihash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32() & 0x7fffffff)
}
