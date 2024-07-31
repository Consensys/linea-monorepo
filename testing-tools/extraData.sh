#!/bin/zsh
# Parallelization factor
N=8

endBlock=$1
blocksBack=$2
dataPoints=$3
endpoint=$4

echo "Args: endBlock=$1 blocksBack=$2 dataPoints=$3 endpoint=$4"

startBlock=$((endBlock - blocksBack))
step=$((blocksBack/dataPoints))
echo "Returning $dataPoints data points from $startBlock to $endBlock with a step of $step"
echo "Scraping data from $endpoint"
for ((i=startBlock; i <= endBlock; i=i+step))
do
  (
  echo -ne "\rProcessing block #$i / $endBlock";
  hexBlockNumber=$(printf "0x%x\n" "$i");
  blockFields=$(curl -sH "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$hexBlockNumber\", false],\"id\":42}" "$endpoint" | jq '. | "\(.result.extraData) \(.result.timestamp) \(.result.number) \(.error.message)"')
  # Removing a quote at the end
  blockFields=${blockFields::-1}
  extraData=$(echo "$blockFields"| cut -d' ' -f 1)
  timestamp=$(echo "$blockFields"| cut -d' ' -f 2)
  blockNumber=$(echo "$blockFields"| cut -d' ' -f 3)
  error=$(echo "$blockFields"| cut -d' ' -f 4)
  decimalTimestamp=$(printf "%d\n" "$timestamp")
  decimalBlockNumber=$(printf "%d\n" "$blockNumber")
  variableCost=$(echo "$extraData" | cut -c 2-66 | cut -c 13-20)
  legacyCost=$(echo "$extraData" | cut -c 2-66 | cut -c 21-28)
  decimalVariableCost=$(printf "%d" "0x$variableCost")
  decimalLegacyCost=$(printf "%d" "0x$legacyCost")
  # File system is used as a hash map pretty much
  echo "$decimalVariableCost $decimalTimestamp $decimalBlockNumber $decimalLegacyCost $error" > "/tmp/$i-extraData.txt" &
  ) &

  if [[ $(jobs -r -p | wc -l) -ge $N ]]; then
    # now there are $N jobs already running, so wait here for any job
    # to be finished so there is a place to start next one.
    wait -n
  fi
done

printf "\nGathering results...\n"
echo "Cleaning output.txt before writing"
rm output.txt
for ((i=startBlock; i <= endBlock; i=i+step))
do
  echo -ne "\rReading /tmp/$i-extraData.txt"
  cat /tmp/"$i"-extraData.txt >> output.txt
  rm /tmp/"$i"-extraData.txt
done
