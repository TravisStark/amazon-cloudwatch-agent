#!/bin/bash

# List of hosts you want to SSH into
HOSTS=(

"monitoring-properties-build-prod-ncl-60001.ncl60.amazon.com"
"lrw-mvprelease-ncl-60001.ncl60.amazon.com"
"lrw-prodfabric-ncl-61001.ncl61.amazon.com"
"monitoring-properties-test-prod-ncl-60001.ncl60.amazon.com"
"cw-snape-ncl-substrate-alpha-60001.ncl60.amazon.com"
"cw-agent-release-ncl-60002.ncl60.amazon.com"
"monitoring-properties-canary-prod-ncl-60002.ncl60.amazon.com"
"lrw-mvprelease-ncl-61001.ncl61.amazon.com"
"lrw-prodfabric-ncl-62001.ncl62.amazon.com"
"lrw-mvprelease-ncl-62001.ncl62.amazon.com"
"lrw-prodfabric-ncl-62003.ncl62.amazon.com"
"monitoring-properties-test-prod-ncl-60002.ncl60.amazon.com"
"lrw-prodfabric-ncl-60002.ncl60.amazon.com"
"monitoring-properties-build-prod-ncl-60002.ncl60.amazon.com"
"monitoring-properties-canary-prod-ncl-60001.ncl60.amazon.com"
"lrw-prodreleasetest-ncl-61001.ncl61.amazon.com"
"lrw-prodfabric-ncl-60001.ncl60.amazon.com"
"lrw-prodfabric-ncl-61002.ncl61.amazon.com"
"cw-agent-release-ncl-s3-60002.ncl60.amazon.com"
"lrw-substratefabric-eqr3zia04wf1.ncl60-1.ec2.substrate"
"lrw-substratereleasetest-ztfvgdn3guk8.ncl61-1.ec2.substrate"
"mon-prop-canary-sub-ncl-23s4vlbpy7rcj.ncl60-1.ec2.substrate"
"mon-prop-test-sub-ncl-2haet9wp5eevo.ncl60-1.ec2.substrate"
"lrw-substratefabric-8sg2d6ilicpj.ncl62-1.ec2.substrate"
"mon-prop-build-sub-ncl-n074swcari6c.ncl60-1.ec2.substrate"
)


# Assign the password for SSH and sudo
PASSWORD="ansh123ansh"

# Remote command to run on each host
REMOTE_COMMAND="sudo yum update -y"

# Log file for tracking progress
LOG_FILE="ssh_update.log"
: > "$LOG_FILE"  # Clear the log file before each run

# Timeout for expect command interactions (adjust if necessary)
TIMEOUT=300  # Increase this if updates take a long time

# Function to run the update command on a single host
run_update_on_host() {
    local HOST=$1

    echo "Connecting to $HOST and running update..." | tee -a "$LOG_FILE"

    # Use expect to handle SSH prompts and password inputs
    expect << EOF | tee -a "$LOG_FILE"
    log_file -a "$LOG_FILE"
    set timeout $TIMEOUT

    spawn ssh $HOST "$REMOTE_COMMAND"

    expect {
        # Handle prompt: "Are you sure you want to continue connecting (yes/no)?"
        "Are you sure you want to continue connecting (yes/no)?" {
            send "yes\r"
            exp_continue
        }
        # Handle passphrase for SSH key
        "Enter passphrase for key" {
            send "$PASSWORD\r"
            exp_continue
        }
        # Handle sudo password prompt
        "assword:" {
            send "$PASSWORD\r"
            exp_continue
        }
        eof {
            send_user "Update completed on $HOST\n"
        }
    }
EOF

    # Check if the expect block executed successfully
    if [ $? -eq 0 ]; then
        echo "Successfully completed update on $HOST" | tee -a "$LOG_FILE"
    else
        echo "Failed to update $HOST" | tee -a "$LOG_FILE"
    fi
}

# Loop over each host and run the update command sequentially
for HOST in "${HOSTS[@]}"; do
    run_update_on_host "$HOST" 
done

echo "Finished updating all hosts." | tee -a "$LOG_FILE"