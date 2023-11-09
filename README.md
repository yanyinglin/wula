# wula
A http performance benchmark with high parallelism and concurrency.

wula is a http performance benchmark tool with high parallelism and concurrency. It is written in Go and can be used to test the performance of any http server.

wula is a Russian word, which means "go rush" in English. It is pronounced as "woo-la". 
wula would be a good choice if you want to test the performance of your http server with high parallelism and concurrency and **IAT** (Inter-Arrival Time) distribution.

wula is for single http url, if you want to test multiple urls, you can use **wulaRR** (wula Round-Robin) instead.

## Citation
wula and wulaRR are used in the following paper:

```
    1. FLASH: Low-Latency Serverless Model Inference with Multi-Core Parallelism in Edge
    
    2. EINS: Edge-Cloud Model Inference with Network-Efficiency Schedule in Serverless 

```