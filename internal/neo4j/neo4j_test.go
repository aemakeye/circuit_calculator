package neo4j

import (
	"bytes"
	"github.com/aemakeye/circuit_calculator/internal/config"
	"github.com/aemakeye/circuit_calculator/internal/storage"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestNeo4jBasic(t *testing.T) {
	var n4jconfig = []byte(`
			{
              "loglevel": "INFO",
			  "neo4j":
			  {
				"host": "localhost",
				"port": "7687",
				"user": "neo4j",
				"password": "password",
				"schema": ""
			  },
			  "minio":
			  {  	
				"host": "localhost",
				"port": "1234",
				"user": "minio",
				"password": "password",
				"schema": ""
			  }
			}
			`)
	tests := []struct {
		name    string
		cypherq string
	}{
		{
			"clear database",
			"CALL apoc.periodic.iterate('MATCH (n) RETURN n', 'DETACH DELETE n', {batchSize:1000})",
		},
		{
			"basic commands",
			"CREATE (a:Greeting) SET a.message = $message RETURN a.message + ', from node ' + id(a)",
		},
		{
			"two nodes creation",
			"CREATE (a:NODE) SET a.uuid = '9cf7c28ccddd4bbebf88a338418c0a6b'" +
				"CREATE (john:Person {name: 'John'})" +
				"CREATE (joe:Person {name: 'Joe'})" +
				"CREATE (steve:Person {name: 'Steve'})" +
				"CREATE (sara:Person {name: 'Sara'})" +
				"CREATE (maria:Person {name: 'Maria'})" +
				"CREATE (john)-[:FRIEND]->(joe)-[:FRIEND]->(steve)" +
				"CREATE (john)-[:FRIEND]->(sara)-[:FRIEND]->(maria)",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := zap.NewNop()
			n4jconfigReader := bytes.NewReader(n4jconfig)
			cfg, err := config.NewConfig(logger, n4jconfigReader)
			assert.NoError(t, err)

			driver, err := neo4j.NewDriver("bolt://"+cfg.Neo4j.Endpoint,
				neo4j.BasicAuth(cfg.Neo4j.User, cfg.Neo4j.Password, ""))
			assert.NoError(t, err)
			defer driver.Close()

			session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
			defer session.Close()

			_, err = session.WriteTransaction(
				func(tx neo4j.Transaction) (interface{}, error) {
					result, err := tx.Run(
						test.cypherq,
						map[string]interface{}{"message": "hello, world"},
					)
					assert.NoError(t, err)

					if result.Next() {
						return result.Record().Values[0], nil
					}

					return nil, err
				},
			)

		})
	}

}

func TestNeo4jController_PushNode(t *testing.T) {
	item := storage.Item{
		UUID:     "eopifrnv-dlfkvn-dklfv",
		ID:       6,
		Value:    "",
		Class:    "resistors",
		SubClass: "resistor_1",
		SourceId: 0,
		TargetId: 0,
		ExitX:    0,
		ExitY:    0,
		EntryX:   0,
		EntryY:   0,
	}
	var n4jconfig = []byte(`
			{
              "loglevel": "INFO",
			  "neo4j":
			  {
				"host": "localhost",
				"port": "7687",
				"user": "neo4j",
				"password": "password",
				"schema": ""
			  }
			}
			`)
	logger := zap.NewNop()
	n4jconfigReader := bytes.NewReader(n4jconfig)
	cfg, err := config.NewConfig(logger, n4jconfigReader)
	assert.NoError(t, err)
	nj, err := NewNeo4j(logger, cfg)
	assert.NoError(t, err)
	t.Run("push node", func(t *testing.T) {
		//uuid, id, err := nj.PushItems(logger, nil)
		assert.NoError(t, err)
		//t.Logf("%s-%s", uuid, id)
	})
}

func TestController_PushItems(t *testing.T) {
}
