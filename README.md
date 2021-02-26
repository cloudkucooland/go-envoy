# go-envoy
Pull data from an Enphase IQ Envoy or IQ Combiner

# This uses the local API, not the Cloud API, so you can poll as frequently as you like
Only basic features will be supported
* reading inventory
* reading current consumption/production/storage data

Tested only with IQ Combiner as that's what I have.

* the endpoints I am requesting are open, but this might be useful
https://thecomputerperson.wordpress.com/2016/08/03/enphase-envoy-s-data-scraping/
https://thecomputerperson.wordpress.com/2016/08/28/reverse-engineering-the-enphase-installer-toolkit/

# TO DO
Most everything.
I'm most curious about /stream/meter
