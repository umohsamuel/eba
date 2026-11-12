Register it:
curl -X POST localhost:8083/connectors -H "Content-Type: application/json" -d @infra/connectors/auth-outbox.json

To change it later, use PUT localhost:8083/connectors/auth-outbox/config with only the config object. To remove it, use DELETE localhost:8083/connectors/auth-outbox.

Check it:
curl localhost:8083/connectors/auth-outbox/status

Both connector.state and tasks[0].state should be RUNNING. If a task shows FAILED, the trace field has the error. The common ones are a wrong password, wal_level not set to logical, or a missing publication. You can also see all of this in kafka-ui at http://localhost:8090, under Kafka Connect.
