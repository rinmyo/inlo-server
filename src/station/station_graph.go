// Copyright ©2017 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package station

import (
	"fmt"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/simple"
)

// StationGraph extends simple.UndirectedGraph to add NewNode and NewEdge
// methods for creating user-defined nodes and edges.
//
// StationGraph implements the encoding.Builder and the dot.Graph
// interfaces.
type StationGraph struct {
	*simple.UndirectedGraph
	id                string
	graph, node, edge attributes
}

// newDotUndirectedGraph returns a new undirected capable of creating user-
// defined nodes and edges.
func NewStationGraph() *StationGraph {
	return &StationGraph{UndirectedGraph: simple.NewUndirectedGraph()}
}

// NewNode adds a new node with a unique node ID to the graph.
func (g *StationGraph) NewNode() graph.Node {
	return &SecNode{Node: g.UndirectedGraph.NewNode()}
}

// NewEdge returns a new Edge from the source to the destination node.
func (g *StationGraph) NewEdge(from, to graph.Node) graph.Edge {
	return &DotEdge{Edge: g.UndirectedGraph.NewEdge(from, to)}
}

// DOTAttributers implements the dot.Attributers interface.
func (g *StationGraph) DOTAttributers() (graph, node, edge encoding.Attributer) {
	return g.graph, g.node, g.edge
}

// DOTUnmarshalerAttrs implements the dot.UnmarshalerAttrs interface.
func (g *StationGraph) DOTAttributeSetters() (graph, node, edge encoding.AttributeSetter) {
	return &g.graph, &g.node, &g.edge
}

// SetDOTID sets the DOT ID of the graph.
func (g *StationGraph) SetDOTID(id string) {
	g.id = id
}

// DOTID returns the DOT ID of the graph.
func (g *StationGraph) DOTID() string {
	return g.id
}

// SecNode extends simple.Node with a label field to test round-trip encoding
// and decoding of node DOT label attributes.
type SecNode struct {
	graph.Node
	dotID string
	// Node label.
	Label string
}

// DOTID returns the node's DOT ID.
func (n *SecNode) DOTID() string {
	return n.dotID
}

// SetDOTID sets a DOT ID.
func (n *SecNode) SetDOTID(id string) {
	n.dotID = id
}

// SetAttribute sets a DOT attribute.
func (n *SecNode) SetAttribute(attr encoding.Attribute) error {
	if attr.Key != "label" {
		return fmt.Errorf("unable to unmarshal node DOT attribute with key %q", attr.Key)
	}
	n.Label = attr.Value
	return nil
}

// Attributes returns the DOT attributes of the node.
func (n *SecNode) Attributes() []encoding.Attribute {
	if len(n.Label) == 0 {
		return nil
	}
	return []encoding.Attribute{{
		Key:   "label",
		Value: n.Label,
	}}
}

type dotPortLabels struct {
	Port, Compass string
}

// DotEdge extends simple.Edge with a label field to test round-trip encoding and
// decoding of edge DOT label attributes.
type DotEdge struct {
	graph.Edge
	// Edge label.
	Label          string
	FromPortLabels dotPortLabels
	ToPortLabels   dotPortLabels
}

// SetAttribute sets a DOT attribute.
func (e *DotEdge) SetAttribute(attr encoding.Attribute) error {
	if attr.Key != "label" {
		return fmt.Errorf("unable to unmarshal node DOT attribute with key %q", attr.Key)
	}
	e.Label = attr.Value
	return nil
}

// Attributes returns the DOT attributes of the edge.
func (e *DotEdge) Attributes() []encoding.Attribute {
	if len(e.Label) == 0 {
		return nil
	}
	return []encoding.Attribute{{
		Key:   "label",
		Value: e.Label,
	}}
}

func (e *DotEdge) SetFromPort(port, compass string) error {
	e.FromPortLabels.Port = port
	e.FromPortLabels.Compass = compass
	return nil
}

func (e *DotEdge) SetToPort(port, compass string) error {
	e.ToPortLabels.Port = port
	e.ToPortLabels.Compass = compass
	return nil
}

func (e *DotEdge) FromPort() (port, compass string) {
	return e.FromPortLabels.Port, e.FromPortLabels.Compass
}

func (e *DotEdge) ToPort() (port, compass string) {
	return e.ToPortLabels.Port, e.ToPortLabels.Compass
}

// attributes is a helper for global attributes.
type attributes []encoding.Attribute

func (a attributes) Attributes() []encoding.Attribute {
	return []encoding.Attribute(a)
}
func (a *attributes) SetAttribute(attr encoding.Attribute) error {
	*a = append(*a, attr)
	return nil
}
