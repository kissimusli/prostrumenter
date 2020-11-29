// This is a package for instrumenting large amout of data programmaticly to prometheus format
// structs require to be a part of the promMetric Interface
// counters and gauges are the two metrics that can be added this way, support for
// histograms and summaries will be added if i see a usercase for it.
// To host a metric, you can use HostMetric function set up your own handlefunction
package prostrumenter

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// errors
var (
	MetricAllreadyExistErr            error = fmt.Errorf("prostrumenter: Metric name allready exist")
	ValueIsANonPointerErr             error = fmt.Errorf("value is a non-pointer should be pointer")
	ValueIsAPointerErr                error = fmt.Errorf("value is a pointer should be non-pointer")
	UpdatetimeCannotBeLessThanZeroErr error = fmt.Errorf("value from getUpdateTime cannot be less than zero")
	ValueShouldBeAstructErr           error = fmt.Errorf("value should be a struct not: ")
)

// the global tag
const tagName = "promstrumenter"

//NewProstrumenter Creates a new Prostrumenter, recommended way of setting up a new Prostrumenter
func NewProstrumenter(Addr, Port string) Prostrumenter {
	a := Prostrumenter{
		Addr: Addr,
		Port: Port,
	}
	return a
}

//GetHandler returns a http handler for hosting the metrics
func (ins *Prostrumenter) GetHandler() http.Handler {
	return promhttp.Handler()
}

//HostMetrics host the instrumented metrics on the prostrumenters addr and port
func (ins *Prostrumenter) HostMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf("%s:%s", ins.Addr, ins.Port), nil)
}

// Instrument a PromMetric pointer, reads the tags of the inputted values and creates a listener that listens to the inputted PromMetric pointers tagged values
// and creates a prometheus counter or gauge The tagged value can be float or int
// tags: `promstrumenter:"counter"` `promstrumenter:"gauge"`
func (ins *Prostrumenter) Instrument(ctx context.Context, p PromMetric) error {

	m := make(map[Field]reflect.Value)
	//map out the struct
	err := mapStruct(m, p)

	if err != nil {
		return err
	}
	//create a listener cointaining counter and so on
	l, err := ins.createListner(p, m)
	if err != nil {
		return err
	}
	//listen
	listen(ctx, l)

	return nil
}

func (ins *Prostrumenter) MultiInstrument(ctx context.Context, p []PromMetric) error {
	for _, prom := range p {
		err := ins.Instrument(ctx, prom)
		if err != nil {
			return err
		}
	}
	return nil
}

//createListners Iterates over the fieldnames and creates a listner for each inputted
func (ins *Prostrumenter) createListner(p PromMetric, m map[Field]reflect.Value) (*listener, error) {
	//create a listener and initiate map
	l := listener{
		counters: make(map[reflect.Value]*counterWrapper),
		gagues:   make(map[reflect.Value]*guageWrapper),
	}

	//update timer cannot be zero or less
	if p.getUpdateTime() < 0 {
		return nil, UpdatetimeCannotBeLessThanZeroErr
	} else {
		l.updateTime = time.Duration(p.getUpdateTime())
	}

	//range over all values and create a counter for each, then sync the wrapper and counter to the value
	for k, v := range m {

		//naming scheme, if the metricname exist generate metricname with uuid
		name := p.getName() + ":" + k.Name
		if ins.exist(name) {
			uuidWithHyphen := uuid.New()
			uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
			name = name + ":" + k.Name + ":" + uuid
		} else {
			ins.MetricNames = append(ins.MetricNames, name)
		}

		switch k.Tag {
		case "counter":
			//create wrapper
			cw := counterWrapper{}
			//better naming struct for multiple use exsist

			//counters
			cw.counter = promauto.NewCounter(prometheus.CounterOpts{
				Name: name,
				Help: p.getHelp(),
			})

			if strings.Contains(k.Type, "int") {
				cw.Type = "int"
				for cw.value < int(v.Int()) {
					cw.value++
					cw.counter.Inc()
				}
			}

			if strings.Contains(k.Type, "float") {
				cw.Type = "float"
				for cw.value < int(v.Float()) {
					cw.value++
					cw.counter.Inc()

				}
			}

			l.counters[v] = &cw

		case "gauge":

			gw := guageWrapper{}

			gw.guage = promauto.NewGauge(prometheus.GaugeOpts{
				Name: name,
				Help: p.getHelp(),
			})

			if strings.Contains(k.Type, "int") {
				gw.Type = "int"
				for gw.value != int(v.Int()) {

					if gw.value < int(v.Int()) {
						gw.value++
						gw.guage.Inc()
					}

					if gw.value > int(v.Int()) {
						gw.value--
						gw.guage.Dec()
					}

				}
			}

			if strings.Contains(k.Type, "float") {
				gw.Type = "float"
				for gw.value != int(v.Float()) {
					if gw.value < int(v.Float()) {
						gw.value++
						gw.guage.Inc()
					}
					if gw.value > int(v.Float()) {
						gw.value--
						gw.guage.Dec()
					}

				}
			}
			l.gagues[v] = &gw

		}

	}

	return &l, nil
}

//Listen sets up a goroutine that updates the values of the counter and the counterwrapper to be equal to the reflected value
func listen(ctx context.Context, l *listener) {
	go func() {
		for {

			time.Sleep(time.Second * l.updateTime)

			for value, counter := range l.counters {
				//check type
				switch counter.Type {
				case "int":
					structValue := int(value.Int())
					for counter.value < structValue {
						counter.counter.Inc()
						counter.value++
					}
				case "float":
					structValue := int(value.Float())
					for counter.value < structValue {
						counter.counter.Inc()
						counter.value++
					}
				}
			}

			for value, guage := range l.gagues {
				//check type
				switch guage.Type {

				case "int":
					structValue := int(value.Int())
					for guage.value != structValue {
						if guage.value < structValue {
							guage.guage.Inc()
							guage.value++
						}
						if guage.value > structValue {
							guage.guage.Dec()
							guage.value--
						}
					}
				case "float":
					structValue := int(value.Float())
					for guage.value != structValue {
						if guage.value < structValue {
							guage.guage.Inc()
							guage.value++
						}
						if guage.value > structValue {
							guage.guage.Dec()
							guage.value--
						}
					}
				}
			}

			//if it is canceled
			select {
			default:
			case <-ctx.Done():
				return
			}
		}
	}()

}

//mapStruct maps an interface and returns a map with values
func mapStruct(m map[Field]reflect.Value, i interface{}) error {
	check := reflect.ValueOf(i)
	if check.Kind() != reflect.Ptr {
		return fmt.Errorf("%v %v", ValueIsAPointerErr, check.Type())
	}
	if check.Kind() != reflect.Struct {
		return fmt.Errorf("%v %v", ValueShouldBeAstructErr, check.Type())
	}

	rv := reflect.ValueOf(i).Elem()
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		valueOf := rv.Field(i)
		typeOf := rt.Field(i)
		fieldType := valueOf.Kind()
		tag := typeOf.Tag.Get(tagName)
		if fieldType == reflect.Struct {

			mapStruct(m, valueOf.Addr().Interface())
		} else {
			//generate uuid
			if tag != "" {
				field := Field{
					Name: typeOf.Name,
					Tag:  tag,
					Type: fieldType.String(),
				}
				m[field] = valueOf
			}

		}

	}

	return nil

}

//generate the metric name
func generateMetricName(p PromMetric, fieldname string) string {
	return fmt.Sprintf("%s:%s", p.getName(), fieldname)
}

//check if the name exist allready
func (ins *Prostrumenter) exist(name string) bool {
	for _, metricname := range ins.MetricNames {
		if metricname == name {
			return true
		}
	}
	return false
}
