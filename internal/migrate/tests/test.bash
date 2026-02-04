cd $(dirname $0)/migrations
rm -f .migrationstate


cleanup() {
    rm -f 0.up.sql
    rm -f .migrationstate
}

trap 'cleanup' EXIT

wantOutput() {
    start=$1
    ending=$2

    if [ $start -le $ending ]; then
        migration_type="up"
        diff=1
    else
        migration_type="down"
        diff=-1
    fi

    i=$start
    while [ $i -ne $ending ]; do
        echo "$migration_type $i"
        i=$(($i + $diff))
    done
}


# Good

if [[ $(go run ../../main.go -command "cat @#" -to 5 .) != $(wantOutput 1 6) ]]
then
    echo "Failed"
    exit 1
fi

if [[ $(go run ../../main.go -command "cat @#" -to 3 .) != $(wantOutput 5 3) ]]
then
    echo "Failed"
    exit 1
fi


if [[ $(go run ../../main.go -command "cat @#" -to TOP .) != $(wantOutput 4 10) ]]
then
    echo "Failed"
    exit 1
fi



if [[ $(go run ../../main.go -command "cat @#" -to 5 .) != $(wantOutput 9 5) ]]
then
    echo "Failed"
    exit 1
fi


if [[ $(go run ../../main.go -command "cat @#" -to BOTTOM .) != $(wantOutput 5 0) ]]
then
    echo "Failed"
    exit 1
fi


if [[ $(go run ../../main.go -command "cat @#" -to TOP .) != $(wantOutput 1 10) ]]
then
    echo "Failed"
    exit 1
fi

if [[ $(go run ../../main.go -command "cat @#" -to BOTTOM .) != $(wantOutput 9 0) ]]
then
    echo "Failed"
    exit 1
fi

# Errors

echo "bad migration" > 0.up.sql


if go run ../../main.go -command "cat @#" -to 5 .
then
    echo "Failed: should have errored because there is a missing down migration"
    exit 1
fi

echo "All tests passed"
