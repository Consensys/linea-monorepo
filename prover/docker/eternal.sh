#!/usr/bin/env bash

# the script is implemented with this logic in mind:
#  (1) blocknumber are written first in the file name and are padded (so we can sort the files by name)
#  (2) the filename is carried over during the process and the extension allows to see where we are
#  (3) we rename the file when we work on it, as a way to take ownership of it
#  (4) we keep the files so we can replay them if necessary
#  (5) we also keep the logs associated to each block

#  Typically, for the block 1387654 we will have:
#     001387654.v123_trace.output_inprogress # named by geth when it creates the file
#     001387654.v123_trace # renamed by geth when the file is done
#     001387654.v123_trace.input_inprogress.4165 # renamed by the script before calling corset
#     001387654.v123_trace.done # renamed by the script when corset has finished

#     001387654.v123_trace.v125_expanded.output_inprogress.4165 # created by corset (name given by the script)
#     001387654.v123_trace.v125_expanded # renamed by the script running corset

#     001387654.v123_trace.v125_expanded.input_inprogress.1657 # renamed by the script running the prover
#     001387654.v123_trace.v125_expanded.done # renamed by the script running the prover

#     001387654.v123_trace.v125_expanded.v123_proof.output_inprogress.1657 # created by the prover (name given by the script)
#     001387654.v123_trace.v125_expanded.v123_proof # renamed by the script running the prover

if [ $# -ne 13 ]; then
    echo Wrong number of argument $#
    echo "Usage: $0 unique_worker_id command_to_run version from_directory input_file_ext to_directory output_file_ext done_directory logs_directory prover-large.sh listing_limit full_prover_extension error_codes"
    echo "Example: $0 $$ cp v123 /tmp/todo_in txt /tmp/todo_out post_cp  /tmp/done /tmp/logs prover-large.sh 100"
    exit
fi

# the unique id of this process. Must be unique between all workers.
LID=$1

# the command to execute. We will call it with the file to process as the last parameter.
CMD=$2

# the version. We used it to generate the filename.
VERSION=$3

# the directory where we read the files to handle
DIR_FROM=$4

# the extension of the file to read (typically trace/expanded/proof)
IN_EXT=$5

# the directory where we write the generated file
DIR_TO=$6

# the extension of the file to create (typically expanded/proof)
OUT_EXT=$7

# the directory where we move the files when they have been handled.
DIR_DONE=$8

# the directory where we write the logs (stdout/stderr)
DIR_LOGS=$9

# the command to execute. We will call it with the file to process as the last parameter.
CMD_LARGE=${10}

LISTING_LIMIT=${11}

# the extension marking a file as requiring a Full-Large prover
LARGE_EXT=${12}

# list of error codes
RETRYABLE_ERROR_CODES=${13}

echo "Unique identifier of the process: $LID"
echo "Command to run: $CMD file_name"
echo "Reading from: $DIR_FROM"
echo "Read extension: $IN_EXT"
echo "Writing to: $DIR_TO"
echo "Write extension: $OUT_EXT"
echo "Moving files \(after execution\) to: $DIR_DONE"
echo "Writing logs to: $DIR_LOGS"
echo "Command to run for prover-large: $CMD_LARGE file_name"
echo "Listing limit: $LISTING_LIMIT"
echo "Error codes: $RETRYABLE_ERROR_CODES"

if [ ! -d "$DIR_FROM" ]; then
    echo "Directory $DIR_FROM does not exist, creating"
    mkdir -p "$DIR_FROM"
fi
if [ ! -d "$DIR_TO" ]; then
    echo "Directory $DIR_TO does not exist, creating"
    mkdir -p "$DIR_TO"
fi
if [ ! -d "$DIR_DONE" ]; then
    echo "Directory $DIR_DONE does not exist, creating"
    mkdir -p "$DIR_DONE"
fi
if [ ! -d "$DIR_LOGS" ]; then
    echo "Directory $DIR_LOGS does not exist, creating"
    mkdir -p "$DIR_LOGS"
fi

# First check if Prover failed and dumped files with error codes
echo Checking if any files were dumped into ${DIR_DONE} from a Prover crash
IFS=',' read -ra retryableErrorCodes <<< "$RETRYABLE_ERROR_CODES"

# We secondly check that we don't have an inprogress file, this can happen if the previous process was stopped.
IP_EXT=inprogress.$LID

echo "Checking if there was a file being processed in $DIR_FROM from a previous run, extension is _$IN_EXT.$IP_EXT"

GEN_IP=$(ls "$DIR_FROM"/*.*"$IN_EXT.$IP_EXT")
if [ ${#GEN_IP} -ge 1 ]; then
    CT=$(echo "$GEN_IP" | wc -l)
    if [ "$CT" -gt 1 ]; then
        echo "ERROR: There should never be more than 1 file being processed when we start, but we found $CT lines, exiting"
        echo files found:
        echo "$GEN_IP"
	    exit
    fi
    LEN=$((${#GEN_IP} - ${#IP_EXT} - 1))
    DEST=$(echo "$GEN_IP" | cut -c 1-$LEN)
    echo "Moving the file that was being processed $GEN_IP to $DEST before continuing"
    mv "$GEN_IP" "$DEST"
else
    echo No files were being processed, continuing
fi

# Used to print a message when nothing happens for a while
NO_FILE_COUNT=0

while true; do
    BEST=""
    LARGE_REQ=0
    # we take the first file, sorted by name, prioritising large files on large provers
    if [[ $LID == *"large"* ]]; then
      LARGE_REQ=1
      BEST=$(find ~ "$DIR_FROM" 2>/dev/null | sort -V | head -n "$LISTING_LIMIT" | grep -i ".*$IN_EXT*" | grep -i ".*$LARGE_EXT$" | head -1)
    fi

    if [ "${BEST}" == "" ]; then
      LARGE_REQ=0
      BEST=$(find ~ "$DIR_FROM" 2>/dev/null | sort -V | head -n "$LISTING_LIMIT" | grep -i ".*$IN_EXT$" | head -1)
    fi

    # check that we found a file
    if [ ${#BEST} -ge 1 ]; then
        NO_FILE_COUNT=0
        BEST_FN=$(basename "$BEST")

        IP_INPUT=$BEST.$IP_EXT
        FN_FINAL_OUTPUT=$BEST_FN.${VERSION}.${OUT_EXT}
        FINAL_OUTPUT=$DIR_TO/$FN_FINAL_OUTPUT
        IP_OUTPUT=$FINAL_OUTPUT.inprogress.$LID
        LOGS=$DIR_LOGS/$FN_FINAL_OUTPUT.logs
        METRICS_FILE=$DIR_LOGS/metrics.txt

        # we rename the file to "own" it. If we fail it means someone else took the ownership.

        if mv "$BEST" "$IP_INPUT" 2>/dev/null; then
            # that's the command we will run. We're just going to log some info
            if [ ${LARGE_REQ} -eq 0 ]; then
              echo Running on Full-Short Prover
              TO_RUN="$CMD $IP_INPUT $IP_OUTPUT"
            else
              echo Running on Full-Large Prover
              TO_RUN="$CMD_LARGE $IP_INPUT $IP_OUTPUT"
            fi
            echo "Found $BEST, and renamed it to $IP_INPUT, going to start $TO_RUN"
            exec_time_in_s="$(date +%s)"
            $TO_RUN &> "$LOGS"
            RES=$?
            exec_time_in_s="$(($(date +%s)-exec_time_in_s))"

            if [ $RES -eq 0 ]; then
                echo "$TO_RUN executed and returned 0, i.e. success, renaming  $IP_OUTPUT to $FINAL_OUTPUT"
                mv "$IP_OUTPUT" "$FINAL_OUTPUT"
                INPUT_DONE="$DIR_DONE"/$BEST_FN.success
            ## --- prover large
            # if failed, run prover-large
            elif [[ " ${retryableErrorCodes[*]} " =~ " ${RES} " ]]; then
              if [[ $LID == *"large"* ]]; then
                echo "Failed with ${RES}, trying to run prover large $IP_OUTPUT to $FINAL_OUTPUT"
                TO_RUN="$CMD_LARGE $IP_INPUT $IP_OUTPUT"
                exec_time_in_s="$(date +%s)"
                $TO_RUN &> "$LOGS"
                RES=$?
                exec_time_in_s="$(($(date +%s)-exec_time_in_s))"
                if [ $RES -eq 0 ]; then
                    echo "$TO_RUN executed and returned 0, i.e. success, renaming  $IP_OUTPUT to $FINAL_OUTPUT"
                    mv "$IP_OUTPUT" "$FINAL_OUTPUT_LARGE"
                    INPUT_DONE="$DIR_DONE"/$BEST_FN.success
                else
                  echo "ERROR: $TO_RUN executed and returned $RES, i.e. an error code, we keep $IP_OUTPUT \(if it exists\) without renaming it"
                  INPUT_DONE="$DIR_DONE"/$BEST_FN.failure.code_$RES
                fi
              else
                echo "ERROR: $TO_RUN executed and returned $RES, move file back to ${DIR_FROM} with ${LARGE_EXT} extension to retry with Large Prover"
                INPUT_DONE="$DIR_FROM"/$BEST_FN.failure.code_$RES.$LARGE_EXT
              fi
              # printing the output of the logs in stdout in case of failure
              cat "$LOGS"
            ## --- prover large
            else
              echo "ERROR: $TO_RUN executed and returned $RES, i.e. an error code, we keep $IP_OUTPUT \(if it exists\) without renaming it"
              INPUT_DONE="$DIR_DONE"/$BEST_FN.failure.code_$RES
              # printing the output of the logs in stdout in case of failure
              cat "$LOGS"
            fi
            echo "${exec_time_in_s}s $INPUT_DONE" >> "$METRICS_FILE"
            echo "${exec_time_in_s} seconds to process $INPUT_DONE"

            mv "$IP_INPUT" "$INPUT_DONE"
        else
            echo "Found $BEST but renaming to $IP_INPUT failed, skipping"
            sleep 1
        fi
    else
        ((NO_FILE_COUNT++))
        IT=$((NO_FILE_COUNT % 100))
        if [ "$IT" -eq 0 ]; then
            echo "No new file after $NO_FILE_COUNT iterations"
        fi
        sleep 1
    fi
done;
