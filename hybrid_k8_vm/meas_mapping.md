# Converting mCAS meas to Prometheus metrics

## DIAM_MEAS

```
diamsch_request_response_avg_milliseconds
	Gauge
    description: Average time interval between the request messages sent and response messages received operations
	label: 
		- type="ALR|CCR|OFR|PMR|RDR|SLR|SNR|SPNR|SRR|STR|TFR|UDR|ACR"
```

```
diamsch_request_response_min_milliseconds	
	Gauge
    description: Minimum time interval between the request messages sent and response messages received operations
	label: 
		- type="ALR|CCR|OFR|PMR|RDR|SLR|SNR|SPNR|SRR|STR|TFR|UDR|ACR"
```

```
diamsch_request_response_max_milliseconds
	Gauge
    description: Maximum time interval between the request messages sent and response messages received operations
	label: 
		- type="ALR|CCR|OFR|PMR|RDR|SLR|SNR|SPNR|SRR|STR|TFR|UDR|ACR"
```

```
diamsch_sent_messages_total
    CounterVec[host string]
    Description: Number diameter messages sent to the remote host
	label: 
		- type="ALR|ALA|CCR|OFR|OFA|PMR|PNA|RDR|RDA|SLA|SNR|SPNA|SRA|SRR|STA|TFA|TFR|UDR|ACR|ACA"
		- code="1XXX|2XXX|3XXX|4XXX|5XXX"
```

```
diamsch_received_messages_total
	CounterVec[host string]
    Description: Number diameter messages received from the remote host
	label: 
		- type="ALR|ALA|CCA|OFR|OFA|PMA|PNR|RDR|RDA|SLR|SNA|SPNR|SRA|SRR|STR|TFA|TFR|UDA|ACR|ACA"
		- code="1XXX|2XXX|3XXX|4XXX|5XXX"
```

Note: The get the total messages gather all messages within a label, otherwise you will get duplicated results
