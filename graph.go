// Package graph is a library for creating generic graph data structures and
// modifying, analyzing, and visualizing them.
//
// # Hashes
//
// A graph consists of vertices of type T, which are identified by a hash value
// of type K. The hash value for a given vertex is obtained using the hashing
// function passed to [New]. A hashing function takes a T and returns a K.
//
// For primitive types like integers, you may use a predefined hashing function
// such as [IntHash] – a function that takes an integer and uses that integer as
// the hash value at the same time:
//
//	g := graph.New(graph.IntHash)
//
// For storing custom data types, you need to provide your own hashing function.
// This example takes a City instance and returns its name as the hash value:
//
//	cityHash := func(c City) string {
//		return c.Name
//	}
//
// Creating a graph using this hashing function will yield a graph of vertices
// of type City identified by hash values of type string.
//
//	g := graph.New(cityHash)
//
// # Operations
//
// Adding vertices to a graph of integers is simple. [graph.Graph.AddVertex]
// takes a vertex and adds it to the graph.
//
//	g := graph.New(graph.IntHash)
//
//	_ = g.AddVertex(1)
//	_ = g.AddVertex(2)
//
// Most functions accept and return only hash values instead of entire instances
// of the vertex type T. For example, [graph.Graph.AddEdge] creates an edge
// between two vertices and accepts the hash values of those vertices. Because
// this graph uses the [IntHash] hashing function, the vertex values and hash
// values are the same.
//
//	_ = g.AddEdge(1, 2)
//
// All operations that modify the graph itself are methods of [Graph]. All other
// operations are top-level functions of by this library.
//
// For detailed usage examples, take a look at the README.
package graph

import (
	"fmt"
	"strings"
)

// Graph represents a generic graph data structure consisting of vertices of
// type T identified by a hash of type K.
type (
	Graph[K comparable, T any] interface {
		GraphRead[K, T]
		GraphWrite[K, T]
	}

	GraphRead[K comparable, T any] interface {
		Hash(T) K
		Traits() Traits

		// Vertex returns the vertex with the given hash or ErrVertexNotFound if it
		// doesn't exist.
		Vertex(hash K) (Vertex[T], error)

		// Vertices returns a slice of all vertices in the graph.
		Vertices() func(yield func(Vertex[T], error) bool)

		// Edge returns the edge joining two given vertices or ErrEdgeNotFound if
		// the edge doesn't exist. In an undirected graph, an edge with swapped
		// source and target vertices does match.
		Edge(sourceHash, targetHash K) (Edge[K], error)

		// Edges returns a slice of all edges in the graph. These edges are of type
		// Edge[K] and hence will contain the vertex hashes, not the vertex values.
		Edges() func(yield func(Edge[K], error) bool)

		// Order returns the number of vertices in the graph.
		Order() (int, error)

		// Size returns the number of edges in the graph.
		Size() (int, error)
	}

	GraphWrite[K comparable, T any] interface {
		// AddVertex creates a new vertex in the graph. If the vertex already exists
		// in the graph, ErrVertexAlreadyExists will be returned.
		//
		// AddVertex accepts a variety of functional options to set further edge
		// details such as the weight or an attribute:
		//
		//	_ = graph.AddVertex("A", "B", graph.VertexWeight(4), graph.VertexAttribute("label", "my-label"))
		//
		AddVertex(value T, options ...func(*VertexProperties)) error

		// UpdateVertex updates the vertex with the given hash value.
		UpdateVertex(hash K, options ...func(*Vertex[T])) error

		// RemoveVertex removes the vertex with the given hash value from the graph.
		//
		// The vertex is not allowed to have edges and thus must be disconnected.
		// Potential edges must be removed first. Otherwise, ErrVertexHasEdges will
		// be returned. If the vertex doesn't exist, ErrVertexNotFound is returned.
		RemoveVertex(hash K) error

		// AddEdge creates an edge between the source and the target vertex.
		//
		// If either vertex cannot be found, ErrVertexNotFound will be returned. If
		// the edge already exists, ErrEdgeAlreadyExists will be returned. If cycle
		// prevention has been activated using PreventCycles and if adding the edge
		// would create a cycle, ErrEdgeCreatesCycle will be returned.
		//
		// AddEdge accepts functional options to set further edge properties such as
		// the weight or an attribute:
		//
		//	_ = g.AddEdge("A", "B", graph.EdgeWeight(4), graph.EdgeAttribute("label", "my-label"))
		//
		AddEdge(sourceHash, targetHash K, options ...func(*EdgeProperties)) error

		// UpdateEdge updates the edge joining the two given vertices with the data
		// provided in the given functional options. Valid functional options are:
		// - EdgeWeight: Sets a new weight for the edge properties.
		// - EdgeAttribute: Adds a new attribute to the edge properties.
		// - EdgeAttributes: Sets a new attributes map for the edge properties.
		// - EdgeData: Sets a new Data field for the edge properties.
		//
		// UpdateEdge accepts the same functional options as AddEdge. For example,
		// setting the weight of an edge (A,B) to 10 would look as follows:
		//
		//	_ = g.UpdateEdge("A", "B", graph.EdgeWeight(10))
		UpdateEdge(source, target K, options ...func(properties *EdgeProperties)) error

		// RemoveEdge removes the edge between the given source and target vertices.
		// If the edge cannot be found, ErrEdgeNotFound will be returned.
		RemoveEdge(source, target K) error
	}

	Vertex[T any] struct {
		Value      T
		Properties VertexProperties
	}

	// Edge represents an edge that joins two vertices. Even though these edges are
	// always referred to as source and target, whether the graph is directed or not
	// is determined by its traits.
	Edge[T any] struct {
		Source     T
		Target     T
		Properties EdgeProperties
	}

	// EdgeProperties represents a set of properties that each edge possesses. They
	// can be set when adding a new edge using the corresponding functional options:
	//
	//	g.AddEdge("A", "B", graph.EdgeWeight(2), graph.EdgeAttribute("color", "red"))
	//
	// The example above will create an edge with a weight of 2 and an attribute
	// "color" with value "red".
	EdgeProperties struct {
		Attributes map[string]string
		Weight     int
		Data       any
	}

	// Hash is a hashing function that takes a vertex of type T and returns a hash
	// value of type K.
	//
	// Every graph has a hashing function and uses that function to retrieve the
	// hash values of its vertices. You can either use one of the predefined hashing
	// functions or provide your own one for custom data types:
	//
	//	cityHash := func(c City) string {
	//		return c.Name
	//	}
	//
	// The cityHash function returns the city name as a hash value. The types of T
	// and K, in this case City and string, also define the types of the graph.
	Hash[K comparable, T any] func(T) K
)

// StringHash is a hashing function that accepts a string and uses that exact
// string as a hash value. Using it as Hash will yield a Graph[string, string].
func StringHash(v string) string {
	return v
}

// IntHash is a hashing function that accepts an integer and uses that exact
// integer as a hash value. Using it as Hash will yield a Graph[int, int].
func IntHash(v int) int {
	return v
}

// EdgeWeight returns a function that sets the weight of an edge to the given
// weight. This is a functional option for the [graph.Graph.Edge] and
// [graph.Graph.AddEdge] methods.
func EdgeWeight(weight int) func(*EdgeProperties) {
	return func(e *EdgeProperties) {
		e.Weight = weight
	}
}

// EdgeAttribute returns a function that adds the given key-value pair to the
// attributes of an edge. This is a functional option for the [graph.Graph.Edge]
// and [graph.Graph.AddEdge] methods.
func EdgeAttribute(key, value string) func(*EdgeProperties) {
	return func(e *EdgeProperties) {
		if e.Attributes == nil {
			e.Attributes = make(map[string]string)
		}
		e.Attributes[key] = value
	}
}

// EdgeAttributes returns a function that sets the given map as the attributes
// of an edge. This is a functional option for the [graph.Graph.AddEdge] and
// [graph.Graph.UpdateEdge] methods.
func EdgeAttributes(attributes map[string]string) func(*EdgeProperties) {
	return func(e *EdgeProperties) {
		e.Attributes = attributes
	}
}

// EdgeData returns a function that sets the data of an edge to the given value.
// This is a functional option for the [graph.Graph.Edge] and
// [graph.Graph.AddEdge] methods.
func EdgeData(data any) func(*EdgeProperties) {
	return func(e *EdgeProperties) {
		e.Data = data
	}
}

// EdgeCopyProperties makes a copy (shallow for .Data) of the given properties and returns a
// 'option'-style function that can be used in the [graph.Graph.AddEdge] and
// [graph.Graph.UpdateEdge] methods.
func EdgeCopyProperties(properties EdgeProperties) func(*EdgeProperties) {
	return func(e *EdgeProperties) {
		if e.Attributes == nil {
			e.Attributes = make(map[string]string)
		}
		for k, v := range properties.Attributes {
			e.Attributes[k] = v
		}
		e.Weight = properties.Weight
		e.Data = properties.Data
	}
}

// EdgeCopy returns the given edge and a function that can be used to copy
// which can be used as arguments to [graph.Graph.AddEdge]
//
//	err := g.AddEdge(EdgeCopy(e))
func EdgeCopy[K comparable](e Edge[K]) (K, K, func(*EdgeProperties)) {
	return e.Source, e.Target, EdgeCopyProperties(e.Properties)
}

// VertexProperties represents a set of properties that each vertex has. They
// can be set when adding a vertex using the corresponding functional options:
//
//	_ = g.AddVertex("A", "B", graph.VertexWeight(2), graph.VertexAttribute("color", "red"))
//
// The example above will create a vertex with a weight of 2 and an attribute
// "color" with value "red".
type VertexProperties struct {
	Attributes map[string]string
	Weight     int
}

// VertexWeight returns a function that sets the weight of a vertex to the given
// weight. This is a functional option for the [graph.Graph.Vertex] and
// [graph.Graph.AddVertex] methods.
func VertexWeight(weight int) func(*VertexProperties) {
	return func(e *VertexProperties) {
		e.Weight = weight
	}
}

// VertexAttribute returns a function that adds the given key-value pair to the
// vertex attributes. This is a functional option for the [graph.Graph.Vertex]
// and [graph.Graph.AddVertex] methods.
func VertexAttribute(key, value string) func(*VertexProperties) {
	return func(e *VertexProperties) {
		if e.Attributes == nil {
			e.Attributes = make(map[string]string)
		}
		e.Attributes[key] = value
	}
}

// VertexAttributes returns a function that sets the given map as the attributes
// of a vertex. This is a functional option for the [graph.Graph.AddVertex] methods.
func VertexAttributes(attributes map[string]string) func(*VertexProperties) {
	return func(e *VertexProperties) {
		e.Attributes = attributes
	}
}

// VertexCopyProperties makes a copy of the given properties and returns
// a 'option'-style function that can be used in the [graph.Graph.AddVertex] and
// [graph.Graph.UpdateVertex] methods.
func VertexCopyProperties(properties VertexProperties) func(*VertexProperties) {
	return func(e *VertexProperties) {
		if e.Attributes == nil {
			e.Attributes = make(map[string]string)
		}
		for k, v := range properties.Attributes {
			e.Attributes[k] = v
		}
		e.Weight = properties.Weight
	}
}

// VertexCopy returns the given vertex and a function that can be used to copy
// which can be used as arguments to [graph.Graph.AddVertex]
//
//	err := g.AddVertex(VertexCopy(v))
func VertexCopy[T any](v Vertex[T]) (T, func(*VertexProperties)) {
	return v.Value, VertexCopyProperties(v.Properties)
}

func EdgesEqual[K comparable, T any](hash Hash[K, T], a, b Edge[T]) bool {
	return hash(a.Source) == hash(b.Source) && hash(a.Target) == hash(b.Target)
}

func CopyTo[K comparable, T any](from GraphRead[K, T], to GraphWrite[K, T]) error {
	for v, err := range from.Vertices() {
		if err != nil {
			return err
		}
		err := to.AddVertex(VertexCopy(v))
		if err != nil {
			return err
		}
	}

	for e, err := range from.Edges() {
		if err != nil {
			return err
		}
		err := to.AddEdge(EdgeCopy(e))
		if err != nil {
			return err
		}
	}

	return nil
}

// EdgeT calls the [graph.Graph.Edge] method and returns the result as an
// [Edge[T]] instead of an [Edge[K]]. This is useful when you want to work with
// the actual vertex values instead of their hash values.
func EdgeT[K comparable, T any](g GraphRead[K, T], source, target K) (Edge[T], error) {
	e, err := g.Edge(source, target)
	if err != nil {
		return Edge[T]{}, err
	}
	sourceV, err := g.Vertex(source)
	if err != nil {
		return Edge[T]{}, fmt.Errorf("failed to get source vertex: %w", err)
	}
	targetV, err := g.Vertex(target)
	if err != nil {
		return Edge[T]{}, fmt.Errorf("failed to get target vertex: %w", err)
	}
	return Edge[T]{
		Source:     sourceV.Value,
		Target:     targetV.Value,
		Properties: e.Properties,
	}, nil
}

func KeysToString[K comparable, T any](g GraphRead[K, T]) (string, error) {
	adj, err := AdjacencyMap(g)
	if err != nil {
		return "", err
	}

	sb := strings.Builder{}
	sb.WriteString("{")
	for src, v := range adj {
		if len(v) == 0 {
			comma := ""
			if sb.Len() > 1 {
				comma = ", "
			}
			fmt.Fprintf(&sb, "%s%v", comma, src)
		}
		for trg := range v {
			comma := ""
			if sb.Len() > 1 {
				comma = ", "
			}
			fmt.Fprintf(&sb, "%s%v -> %v", comma, src, trg)
		}
	}
	sb.WriteString("}")
	return sb.String(), nil
}
