#! /bin/bash
TEST_OVERWRITTEN_BY_ENV=env TEST_OVERWRITTEN_BY_CONF_THEN_ENV=env TEST_OVERWRITTEN_BY_CONF_THEN_ENV_THEN_ARG=env go run main.go --overwritten-by-conf-then-env-then-arg arg --config c.conf --overwritten-by-conf-then-arg arg

if [ $? -ne 0 ]; then
echo "Test failed"
else
echo "Test passed"
fi