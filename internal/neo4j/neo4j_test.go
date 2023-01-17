package neo4j

import (
	"fmt"
	"github.com/aemakeye/circuit_calculator/internal/drawio"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

const (
	url      = "127.0.0.1:7687"
	user     = "neo4j"
	password = "password"
)

var ctrlr, _ = NewController(zap.NewNop(), url, user, password)

func TestNeo4jBasic(t *testing.T) {

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

			driver := ctrlr.Driver
			session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
			defer session.Close()

			_, err := session.WriteTransaction(
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
			assert.NoError(t, err)

		})
	}

}

func TestNeo4jController_PushNode(t *testing.T) {

	var input []drawio.Item
	for i := 1; i < 9; i++ {
		input = append(input, drawio.Item{
			UUID:     "eopifrnv-dlfkvn-dklfv",
			EID:      i,
			Class:    "resistors",
			SubClass: "resistor_1",
		})
	}

	logger := zap.NewNop()
	ichan := make(chan drawio.Item)
	rchan := make(chan pushResult)

	t.Run("push node, expected 8", func(t *testing.T) {
		pushedNodesCount := 0
		go ctrlr.PushNodes(logger, ichan, rchan)

		go func(tt *testing.T, rch chan pushResult) {
			for v := range rch {
				assert.NoError(tt, v.error)
			}
		}(t, rchan)

		for _, v := range input {
			ichan <- v
			pushedNodesCount++
		}

		close(ichan)
		close(rchan)
		assert.Equal(t, 8, pushedNodesCount)
	})
}

func TestController_PushRelation(t *testing.T) {
	//TODO: tests do not work if run by single button. FIX this. now please run one by one
	logger := zap.NewNop()
	var nodeInput, rinput []drawio.Item
	for i := 1; i < 9; i++ {
		nodeInput = append(nodeInput, drawio.Item{
			UUID:     "eopifrnv-dlfkvn-dklfv",
			EID:      i,
			Class:    drawio.ItemClassResistors,
			SubClass: "resistor_1",
		})
	}

	for i := 0; i < 9; i++ {
		rinput = append(rinput, drawio.Item{
			UUID:     "eopifrnv-dlfkvn-dklfv",
			EID:      0,
			Value:    "",
			Class:    drawio.ItemClassLines,
			SubClass: "",
			SourceId: i,
			TargetId: i + 1,
			ExitX:    0,
			ExitY:    0,
			EntryX:   0,
			EntryY:   0,
		})
	}

	ichan := make(chan drawio.Item)
	relchan := make(chan drawio.Item)
	reschan := make(chan pushResult)
	reschan2 := make(chan pushResult)

	tests := []struct {
		Name          string
		ExpectedError error
		Relations     []drawio.Item
		Reschan       chan pushResult
	}{
		{
			"missing node for relation",
			fmt.Errorf("neo4j transaction failed, empty return value"),
			[]drawio.Item{rinput[0]},
			reschan,
		},
		{
			"all good",
			nil,
			rinput[1:],
			reschan2,
		},
	}

	go func(tt *testing.T, rch chan pushResult) {
		for v := range rch {
			assert.NoError(tt, v.error)
		}
	}(t, reschan)

	go func(chan drawio.Item) {
		ctrlr.PushNodes(logger, ichan, reschan)
	}(ichan)

	for _, item := range nodeInput {
		t.Logf("pushing node id: %d", item.EID)
		ichan <- item
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			//check pushResult.error value
			go func(tt *testing.T, rch chan pushResult) {
				for v := range rch {
					if test.ExpectedError != nil {
						t.Logf("entered error branch")
						assert.EqualError(tt, v.error, test.ExpectedError.Error())
					} else {
						assert.NoError(tt, v.error)
					}

				}

			}(t, test.Reschan)

			go func(chan drawio.Item) {
				ctrlr.PushRelation(logger, relchan, reschan)
			}(relchan)

			for _, item := range test.Relations {
				t.Logf("pushing relation with source %d and dest %d", item.SourceId, item.TargetId)
				relchan <- item
			}
		})
	}
	close(ichan)
	close(relchan)
	close(reschan)
	//close(reschan2)
}
