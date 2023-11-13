package main

import (
	"log"
	"net/http"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Structure with the contents of a meas table row
type MeasEntry struct {
	name string
	value float64
}

func (m MeasEntry) String() string {
	return fmt.Sprintf("MeasEntry{name: %s, value: %f}", m.name, m.value,)
}

// Definition of a metric made from a meas reading:
// - desc, is a prometheus.Desc constant with the Metric description
// - type, is the Metric type enum value CounterValue, GaugeValue
// - label, the same meas table entry can be mapped to label(s) 
// - value, the reading value
type MeasMetric struct {	
	metric_type prometheus.ValueType
	label string 
	desc *prometheus.Desc
	value float64
}

func (m MeasMetric) String() string {
	return fmt.Sprintf("MeasMetric{metric_type:%d, label:%s, desc:%s, value:%f", m.metric_type, m.label, m.desc, m.value)
}

// Interface for populating structured MeasEntry from meas data source
type MeasFetcher interface {
	// Goes to the data source (e.g. db) and returns a list of meas entries
	GetMeasEntries() []MeasEntry
}

// Implementation for diamsch_meas table
// Is this definition really needed?
type DiameterMeasFetcher struct {
}

func (diameterFetcher DiameterMeasFetcher) GetMeasEntries() []MeasEntry {
	//tableData := getDataFromTable("diamsch_meas")
	// get name and Collation and map it to MeasEntry
	
	// Build sample data
	measEntries := make([]MeasEntry, 8, 8)
	measEntries[0] = MeasEntry { name: "avg_tm_ccr_rsp", value: 3000 }
	measEntries[1] = MeasEntry { name: "min_tm_ccr_rsp", value: 2000 }
	measEntries[2] = MeasEntry { name: "max_tm_ccr_rsp", value: 4000 }
	measEntries[3] = MeasEntry { name: "num_ccr_sent", value: 100 }
	measEntries[4] = MeasEntry { name: "num_udr_sent", value: 30 }
	measEntries[5] = MeasEntry { name: "num_cca_rcvd", value: 100 }
	measEntries[6] = MeasEntry { name: "num_uda_rcvd", value: 30 }
	// measEntries[7] = MeasEntry { name: "num_succ_2xxx_rcvd", value: 120 }
	// measEntries[8] = MeasEntry { name: "num_err_3xxx_rcvd", value: 10 }
	measEntries[7] = MeasEntry { name: "so_para_ver_se_chora", value: 10 }

	return measEntries
}

// Descriptors used by the DiameterMeasConverter and MeasCollector below
// Make these vars to avoid having to deal with ownership and single initialization
var (
	diamAvgResDesc = prometheus.NewDesc(
		"diamsch_avg_response_milliseconds",
		"Average time interval between the request messages sent and response messages received operations.",
		[]string{"type"}, nil,
	)
	diamMinResDesc = prometheus.NewDesc(
		"diamsch_min_response_milliseconds",
		"Minimum time interval between the request messages sent and response messages received operations.",
		[]string{"type"}, nil,
	)
	diamMaxResDesc = prometheus.NewDesc(
		"diamsch_max_response_milliseconds",
		"Maximum time interval between the request messages sent and response messages received operations.",
		[]string{"type"}, nil,
	)	
	diamSentMsg = prometheus.NewDesc(
		"diamsch_sent_messages_total",
		"Number diameter messages sent to the remote host.",
		[]string{"type"}, nil,
	)
	diamRcvdMsg = prometheus.NewDesc(
		"diamsch_received_messages_total",
		"Number diameter messages received from the remote host.",
		[]string{"type"}, nil,
	)
)

// Each meas table implements a MeasConverter
// It retrieves all the Metrics associated to a meas table 
type MeasConverter interface {
	// Creates the MeasMetric from the given meas table entry
	CreateMetricFromEntry(entry MeasEntry) *MeasMetric 
	// Retrieves the Metrics definition and value associated to this table
	CreateMetricsFromTable() *[]MeasMetric
}

//  Converts diamsch_meas meas to Metrics
type DiameterMeasConverter struct {
	diamFetcher DiameterMeasFetcher
	
}

func (measConv DiameterMeasConverter) CreateMetricFromEntry(entry MeasEntry) *MeasMetric {		
	fmt.Printf("Creating MeasMetric from %s\n", entry)
	
	switch entry.name {
	case "num_ccr_sent":
		return &MeasMetric{ 
			desc: diamSentMsg, 
			label: "CCR", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	case "num_cca_rcvd":
		return &MeasMetric{ desc: diamRcvdMsg, 
			label: "CCA", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	case "num_udr_sent":
		return &MeasMetric{ desc: diamSentMsg, 
			label: "CCR", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	case "num_uda_rcvd":
		return &MeasMetric{ desc: diamRcvdMsg, 
			label: "CCA", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	case "avg_tm_ccr_rsp":
		return &MeasMetric{ desc: diamAvgResDesc, 
			label: "CCR", 
			value: entry.value, 
			metric_type: prometheus.GaugeValue }
	case "avg_tm_udr_rsp":
		return &MeasMetric{ desc: diamAvgResDesc, 
			label: "UDR", 
			value: entry.value, 
			metric_type: prometheus.GaugeValue }
	case "min_tm_ccr_rsp":
		return &MeasMetric{ desc: diamMinResDesc, 
			label: "CCR", 
			value: entry.value, 
			metric_type: prometheus.GaugeValue }
	case "min_tm_udr_rsp":
		return &MeasMetric{ desc: diamMinResDesc, 
			label: "UDR", 
			value: entry.value, 
			metric_type: prometheus.GaugeValue }
	case "max_tm_ccr_rsp":
		return &MeasMetric{ desc: diamMaxResDesc, 
			label: "CCR", 
			value: entry.value, 
			metric_type: prometheus.GaugeValue }
	case "max_tm_udr_rsp":
		return &MeasMetric{ desc: diamMaxResDesc, 
			label: "UDR", 
			value: entry.value, 
			metric_type: prometheus.GaugeValue }
	case "num_info_1xxx_rcvd":
		return &MeasMetric{ desc: diamRcvdMsg, 
			label: "1xxx", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	case "num_succ_2xxx_rcvd":
		return &MeasMetric{ desc: diamRcvdMsg, 
			label: "2xxx", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	case "num_err_3xxx_rcvd":
		return &MeasMetric{ desc: diamRcvdMsg, 
			label: "3xxx", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	case "num_tran_4xxx_rcvd":
		return &MeasMetric{ desc: diamRcvdMsg, 
			label: "4xxx", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	case "num_perm_5xxx_rcvd":
		return &MeasMetric{ desc: diamRcvdMsg, 
			label: "5xxx", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	case "num_info_1xxx_sent":
		return &MeasMetric{ desc: diamSentMsg, 
			label: "1xxx", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	case "num_succ_2xxx_sent":
		return &MeasMetric{ desc: diamSentMsg, 
			label: "2xxx", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	case "num_err_3xxx_sent":
		return &MeasMetric{ desc: diamSentMsg, 
			label: "3xxx", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	case "num_tran_4xxx_sent":
		return &MeasMetric{ desc: diamSentMsg, 
			label: "4xxx", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	case "num_perm_5xxx_sent":
		return &MeasMetric{ desc: diamSentMsg, 
			label: "5xxx", 
			value: entry.value, 
			metric_type: prometheus.CounterValue }
	}

	return nil
}

func (measConv DiameterMeasConverter) CreateMetricsFromTable() *[]MeasMetric {	
	//assumes all entries have values, though these may not match the converter rules
	entries := measConv.diamFetcher.GetMeasEntries()	
	
	metrics := make([]MeasMetric, len(entries), cap(entries))	
	n := 0
	for i, entry := range entries {
		m := measConv.CreateMetricFromEntry(entry)

		// When no metric is matched (nil), skip entry
		// count non empty metrics
		if m != nil {
			fmt.Printf("Append metric %s \n", m)
			metrics[i] = *m
			n++
		}		
	}	

	metrics = metrics[:n]	
	return &metrics
}

// Implements the Collector interface
// Creates on the fly metrics for all that is returned by each MeasConverter
type MeasCollector struct {
	measConverters []MeasConverter
}

func (measColl MeasCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(measColl, ch)
}

func (mc MeasCollector) Collect(ch chan<- prometheus.Metric) {
	for _, converter := range mc.measConverters {		
		for _, metric := range *converter.CreateMetricsFromTable() {
			ch <- prometheus.MustNewConstMetric(
				metric.desc,
				metric.metric_type,
				metric.value,				
				metric.label,
			)
		}
	}
}

func NewMeasCollector() *MeasCollector {
	// create a Diameter converter that uses the DiameterMeasFetcher as source
	diamConverter := DiameterMeasConverter{ diamFetcher: DiameterMeasFetcher{} }
	
	// add Diameter converter to the list of coverters
	converters := []MeasConverter { diamConverter }

	// init the Meas collector with the existing data converters
	collector :=  MeasCollector { measConverters: converters }

	return &collector
}

func main() {
	reg := prometheus.NewPedanticRegistry()

	collector := NewMeasCollector()
	prometheus.MustRegister(collector)
		
	reg.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)

	// Expose metrics and custom registry via an HTTP server
	// using the HandleFor function. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
