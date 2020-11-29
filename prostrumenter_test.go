package prostrumenter

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

type Test struct {
	Namespace          string
	Name               string
	Help               string
	UpdateTime         int
	nonExported        int     `promstrumenter:"counter"`
	TestCounterInt     int     `promstrumenter:"counter"`
	TestCounterInt8    int8    `promstrumenter:"counter"`
	TestCounterInt16   int16   `promstrumenter:"counter"`
	TestCounterInt32   int32   `promstrumenter:"counter"`
	TestCounterInt64   int64   `promstrumenter:"counter"`
	TestCounterFloat32 float32 `promstrumenter:"counter"`
	TestCounterFloat64 float64 `promstrumenter:"counter"`
	TestGuageInt       int     `promstrumenter:"gauge"`
	TestGuageInt8      int8    `promstrumenter:"gauge"`
	TestGuageInt16     int16   `promstrumenter:"gauge"`
	TestGuageInt32     int32   `promstrumenter:"gauge"`
	TestGuageInt64     int64   `promstrumenter:"gauge"`
	TestGuageFloat32   float32 `promstrumenter:"gauge"`
	TestGuageFloat64   float64 `promstrumenter:"gauge"`
}

func (n Test) getName() string {
	return n.Name
}
func (n Test) getHelp() string {
	return n.Help
}
func (n Test) getUpdateTime() int {
	return n.UpdateTime
}

type Nested struct {
	nonExported     int `promstrumenter:"counter"`
	ExportedNoTag   int
	TestCounterInt  int  `promstrumenter:"counter"`
	TestCounterInt8 int8 `promstrumenter:"counter"`
	Nest            NestedLevel1
	MassNested      []NestedLevel1
}

type NestedLevel1 struct {
	nonExported   int `promstrumenter:"counter"`
	ExportedNoTag int
	TestGuageInt  int  `promstrumenter:"gauge"`
	TestGuageInt8 int8 `promstrumenter:"gauge"`
	Nest          NestedLevel2
}

type NestedLevel2 struct {
	nonExported      int `promstrumenter:"counter"`
	ExportedNoTag    int
	TestGuageFloat32 float32 `promstrumenter:"gauge"`
	TestGuageFloat64 float64 `promstrumenter:"gauge"`
}

var (
	Addr = ""
	Port = "41112"
)

//TestInstrumenting
func TestInstrumenting(t *testing.T) {
	//testcases
	test := Test{
		Namespace:        "Ninjas",
		Name:             "Ninja",
		Help:             "BengtsData",
		TestCounterInt:   0,
		TestGuageFloat32: 500,
		UpdateTime:       1,
	}
	ctx, cancel := context.WithCancel(context.Background())

	//create a new Postrumenter
	p := NewProstrumenter(Addr, Port)
	//instrument our struct
	err := p.Instrument(ctx, &test)
	err = p.Instrument(ctx, &test)
	if err != nil {
		t.Errorf("%v", err)
		cancel()
		return
	}

	//Change the stats of the testcase
	go statChange(&test)

	p.HostMetrics()
	cancel()
}

//TestMultiInstrument
func TestMultiInstrument(t *testing.T) {
	p := NewProstrumenter(Addr, Port)
	ctx, cancel := context.WithCancel(context.Background())
	test := Test{
		Namespace:        "Ninjas",
		Name:             "Ninja",
		Help:             "Contains every type of data",
		TestCounterInt:   0,
		TestGuageFloat32: 500,
		UpdateTime:       1,
	}
	test2 := Test{
		Namespace:        "Ninjas",
		Name:             "Ninja",
		Help:             "Contains every type of data",
		TestCounterInt:   0,
		TestGuageFloat32: 500,
		UpdateTime:       1,
	}
	go statChange(&test)
	go statChange(&test2)

	a := []PromMetric{}

	a = append(a, &test)
	a = append(a, &test2)

	err := p.MultiInstrument(ctx, a)
	if err != nil {
		t.Errorf("%v", err)
		cancel()
		return
	}

	p.HostMetrics()
	cancel()
}

func TestMapStruct(t *testing.T) {
	/* 	p := NewProstrumenter(Addr, Port) */

	test := Nested{
		nonExported:     1,
		ExportedNoTag:   2,
		TestCounterInt:  3,
		TestCounterInt8: 4,
		Nest: NestedLevel1{
			nonExported:   11,
			ExportedNoTag: 12,
			TestGuageInt:  13,
			TestGuageInt8: 14,
			Nest: NestedLevel2{
				nonExported:      21,
				ExportedNoTag:    22,
				TestGuageFloat32: 23,
				TestGuageFloat64: 24,
			},
		},
	}

	m := make(map[Field]reflect.Value)

	err := mapStruct(m, &test)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v\n", m)

	for k, v := range m {
		fmt.Printf("Field: %v", k)
		fmt.Printf("Value: %v\n ", v)
	}
	test.TestCounterInt = 42

	for k, v := range m {
		fmt.Printf("Field: %v", k)
		fmt.Printf("Value: %v\n ", v)
	}
}

func TestCreateListner(t *testing.T) {
	p := NewProstrumenter(Addr, Port)
	m := make(map[Field]reflect.Value)
	test := Test{
		Namespace:        "Ninjas",
		Name:             "Ninja",
		Help:             "BengtsData",
		TestCounterInt:   0,
		TestGuageFloat32: 500,
	}
	err := mapStruct(m, &test)
	if err != nil {
		t.Error(err)
	}

	l, err := p.createListner(&test, m)
	if err != nil {
		t.Error(err)
	}

	for _, c := range l.counters {
		fmt.Printf("\n %v", c.Type)
	}
}

func BenchmarkInstrument(b *testing.B) {

	test := Nested{
		nonExported:     1,
		ExportedNoTag:   2,
		TestCounterInt:  3,
		TestCounterInt8: 4,
		Nest: NestedLevel1{
			nonExported:   11,
			ExportedNoTag: 12,
			TestGuageInt:  13,
			TestGuageInt8: 14,
			Nest: NestedLevel2{
				nonExported:      21,
				ExportedNoTag:    22,
				TestGuageFloat32: 23,
				TestGuageFloat64: 24,
			},
		},
	}

	m := make(map[Field]reflect.Value)

	err := mapStruct(m, &test)
	if err != nil {
		b.Error(err)
	}
}

func statChange(nin *Test) {
	for {
		time.Sleep(2 * time.Second)

		nin.TestCounterFloat32 = nin.TestCounterFloat32 + 2.5
		nin.TestCounterFloat64 = nin.TestCounterFloat64 + 2.5
		nin.TestGuageFloat32 = float32(rand.Intn(100)-50) + rand.Float32()
		nin.TestGuageFloat64 = float64(rand.Intn(100)-50) + rand.Float64()

		nin.TestCounterInt = nin.TestCounterInt + 2
		nin.TestCounterInt8 = int8(nin.TestCounterInt8 + 3)
		nin.TestCounterInt16 = int16(nin.TestCounterInt16 + 3)
		nin.TestCounterInt32 = int32(nin.TestCounterInt32 + 3)
		nin.TestCounterInt64 = int64(nin.TestCounterInt64 + 3)

		nin.TestGuageInt = (rand.Intn(100) - 50)
		nin.TestGuageInt8 = int8(rand.Intn(100) - 50)
		nin.TestGuageInt16 = int16(rand.Intn(100) - 50)
		nin.TestGuageInt32 = int32(rand.Intn(100) - 50)
		nin.TestGuageInt64 = int64(rand.Intn(100) - 50)

	}
}
