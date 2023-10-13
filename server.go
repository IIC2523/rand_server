package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Node struct {
	Name           string
	Number         int
	Files          []string
	Subdirectories []string
	Left           *Node
	Right          *Node
}

var number int

type Response struct {
	Files          []string `json:"files"`
	Subdirectories []string `json:"subdirectories"`
}

func buildRandTree(depth int, lucky int) *Node {
	node := new(Node)
	node.Number = number
	if number == lucky {
		node.Files = append(node.Files, "gopher.jpg")
	} else {
		for i := 0; i < rand.Intn(10); i++ {
			node.Files = append(node.Files, "file"+strconv.Itoa(i)+strconv.Itoa(number)+".txt")
		}
	}
	if number == 0 {
		node.Name = "index"
	} else {
		node.Name = "d" + strconv.Itoa(number)
	}
	if depth == 0 {
		return node
	}
	// Build left
	number++
	node.Left = buildRandTree(depth-1, lucky)
	if node.Left != nil {
		node.Subdirectories = append(node.Subdirectories, node.Left.Name)
	}

	// Build right
	number++
	node.Right = buildRandTree(depth-1, lucky)
	if node.Right != nil {
		node.Subdirectories = append(node.Subdirectories, node.Right.Name)
	}
	return node
}

func treeToHashMap(node *Node, hashMap map[string]Response) map[string]Response {
	if node == nil {
		return hashMap
	}
	var data Response
	data.Files = node.Files
	data.Subdirectories = node.Subdirectories
	hashMap[node.Name] = data
	hashMap = treeToHashMap(node.Left, hashMap)
	hashMap = treeToHashMap(node.Right, hashMap)
	return hashMap
}

func storeHashMap(hashMap map[string]Response) {
	file, err := os.Create("hashMap.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	for key, value := range hashMap {
		data := fmt.Sprintf("%v", value)
		jsonData, err := json.Marshal(value)
		if err != nil {
			fmt.Println(err)
			return
		}
		data = string(jsonData)
		line := key + "&" + data + "\n"
		_, err = file.WriteString(line)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func loadHashMap() map[string]Response {
	file, err := os.Open("hashMap.txt")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()
	hashMap := make(map[string]Response)
	var key string
	var data string
	var value Response
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, "&")
		key = split[0]
		data = split[1]
		err := json.Unmarshal([]byte(data), &value)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		hashMap[key] = value
		key = ""
		data = ""
		value = Response{}
	}
	return hashMap
}

func listener(port int, treeMap map[string]Response) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.Path[1:]
		fmt.Println(uri)
		if uri[len(uri)-1:] == "/" {
			uri = uri[:len(uri)-1]
		}
		split := strings.Split(uri, "/")
		uri = split[len(split)-1]
		var response Response
		response = treeMap[uri]
		enc := json.NewEncoder(w)
		enc.Encode(response)
	})
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func powInt(x, y int) int {
	return int(math.Pow(float64(x), float64(y)))
}

func main() {

	var mode, depth, nodes, port int

	flag.IntVar(&mode, "m", 0, "Mode: 0 for build, 1 for serve")
	flag.IntVar(&depth, "d", 0, "Depth of tree")
	flag.IntVar(&port, "p", 8080, "Port to listen on")

	flag.Parse()

	nodes = powInt(2, depth+1) - 3

	if mode == 0 {
		fmt.Println("Building tree...")
		number = 0
		lucky := rand.Intn(nodes)
		tree := buildRandTree(depth, lucky)
		hashMap := make(map[string]Response)
		hashMap = treeToHashMap(tree, hashMap)
		storeHashMap(hashMap)
		fmt.Println("Tree built")
		return
	} else if mode == 1 {
		fmt.Println("Launching server on port " + strconv.Itoa(port) + "...")
		treeMap := loadHashMap()
		listener(port, treeMap)
		return
	} else {
		fmt.Println("Invalid mode")
		return
	}

}
