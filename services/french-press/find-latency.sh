#!/usr/bin/env bash
kafkaEndpoint=http://$kafkaEndpoint/topics/voltron.Latency
filename="remote-destinations.txt"
while true
do
    while read -r line
    do
        name="$line"
        interfaces_file="interfaces.txt"
        while read -r line
        do
            interface=$line
            IFS=
            var=$(ping -c5 -q -s1400 -w1200 -I $interface $name)
            ping_statistics=$(echo $var | tr "\n" "\n")
            line=$(echo -e $ping_statistics | sed -n '5p')
            N=4
            var="$(echo $line | cut -d " " -f $N)"
            var="$(echo $var | cut -d'/' -f2)"
            time="$var"
            echo $time
            var=$(curl -X POST -H "Content-Type: application/vnd.kafka.json.v2+json" --data '{"records":[{"value":{"from_ip":"'$interface'","to_ip":"'$name'","latency":"'$time'"}}]}' $kafkaEndpoint)
            echo $var
        done < "$interfaces_file"
    done < "$filename"
    sleep 300
done
