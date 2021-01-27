package main

import (
	"fmt"
	edot "gonum.org/v1/gonum/graph/encoding/dot"
	fdot "gonum.org/v1/gonum/graph/formats/dot"
	"pracserver/src/station"
)

func main() {
	ast, err := fdot.ParseFile("./resource/stations/linli/relation_s.dot")
	if err != nil {
		fmt.Println("解析文件錯誤: ", err)
	}
	astGraph := ast.Graphs[0]
	graph := station.NewStationGraph()
	err = edot.Unmarshal([]byte(astGraph.String()), graph)
	if err != nil {
		fmt.Println("解析圖錯誤: ", err)
	}
	fmt.Println(graph.DOTID(), ": ")

	nodes := graph.Nodes()
	for nodes.Next() {
		node := nodes.Node().(*station.SecNode)
		println(node.DOTID(), ": ")
		for _, attr := range node.Attributes() {
			print(attr.Key, "=", attr.Value, " ")
		}
	}

	edges := graph.Edges()
	for edges.Next() {
		edge := edges.Edge().(*station.DotEdge)
		println(edge.From().(*station.SecNode).DOTID(), "--", edge.To().(*station.SecNode).DOTID())
	}
}
