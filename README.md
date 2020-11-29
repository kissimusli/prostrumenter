# prostrumenter
A easy way to instrument structs and export them as a prometheus metric type.

This is a package for instrumenting large amout of data programmaticly to the prometheus metric type.

How To use:

Structs must be a part of the PromMetric interface and have the tags: `promstrumenter:"counter"` or `promstrumenter:"gauge"` on the metrics that you wish to instrument.

Postrumenter will create a goroutine that tracks the metric and updates counters and gauges.

To host a metric, you can use HostMetric function set up your own handlefunction and get the handler from the program.

Histograms and summaries will be added if i see a usercase for it.

