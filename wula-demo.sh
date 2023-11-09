declare timestamp=`date +%s%3N`
python gendist.py -name wula -dir ./wula-examples -t 1 -mu 0.1 -sigma 0.01
./wula/wula -name wula1 -dist wula -dir ./wula-examples -url http://127.0.0.1:8080/ -dest ./wula-examples/wula1-request.csv -SLO 10 -synctime $(expr $timestamp + 500) &
./wula/wula -name wula2 -dist wula -dir ./wula-examples -url http://127.0.0.1:8080/ -dest ./wula-examples/wula2-request.csv -SLO 10 -synctime $(expr $timestamp + 1000) &
wait