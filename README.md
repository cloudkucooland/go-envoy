# go-envoy
Pull data from an Enphase IQ Envoy or IQ Combiner

# This uses the local API, not the Cloud API, so you can poll as frequently as you like
Only basic features will be supported
* reading inventory
* reading current consumption/production/storage data

# examples
* see current production / max, consumption, net statistics
* max is calculated based on the sum of the highest production each panel has produced since last IQ reboot
```
% envoy now
Production: 6698.88W / 8808W	Consumption: 722.03W	Net: -5976.85W
```

* see today's totals
```
% envoy today
Production: 13.55kWh	Consumption: 8.58kWh	Net: 0.00kWh
```

* see envoy info
```
% envoy info
Serial Number:  xxx
Part Number:  800-00555-r03
Software Version:  R4.10.35
```

* Tested only with IQ Combiner as that's what I have.

* the endpoints I am requesting are open, but this might be useful
https://thecomputerperson.wordpress.com/2016/08/03/enphase-envoy-s-data-scraping/
https://thecomputerperson.wordpress.com/2016/08/28/reverse-engineering-the-enphase-installer-toolkit/

# TO DO
This does what I need, if there are features you'd like, please let me know
