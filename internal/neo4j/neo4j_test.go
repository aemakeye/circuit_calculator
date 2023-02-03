package neo4j

import (
	"fmt"
	"github.com/aemakeye/circuit_calculator/internal/drawio"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"math/rand"
	_ "net/http/pprof"
	"sync"
	"testing"
	"time"
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
			defer func() {
				_ = session.Close()
			}()

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
	rchan := make(chan drawio.Item)

	t.Run("push node, expected 8", func(t *testing.T) {
		pushedNodesCount := 0
		noMoreNodes := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			ctrlr.PushNodes(logger, ichan, rchan, noMoreNodes)
			wg.Done()
		}()

		for _, v := range input {
			ichan <- v
			r, ok := <-rchan
			if ok {
				pushedNodesCount++
			}
			t.Logf("result id %d", r.EID)

			//time.Sleep(1 * time.Second)
		}
		noMoreNodes <- struct{}{}

		wg.Wait()
		assert.Equal(t, len(input), pushedNodesCount)
	})
}

func TestController_PushRelation(t *testing.T) {
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

	for i := 0; i < 8; i++ {
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
	reschan := make(chan drawio.Item)
	reschan2 := make(chan drawio.Item)

	tests := []struct {
		Name          string
		ExpectedError error
		Relations     []drawio.Item
		Reschan       chan drawio.Item
	}{
		{
			"missing node for relation",
			fmt.Errorf("neo4j transaction failed, empty return value for relation"),
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

	go ctrlr.PushNodes(logger, ichan, reschan, nil)

	for _, item := range nodeInput {
		t.Logf("pushing node id: %d", item.EID)
		ichan <- item
		res := <-reschan
		assert.NoError(t, res.Error)
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			noMoreRels := make(chan struct{})
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				ctrlr.PushRelations(logger, relchan, reschan, noMoreRels)
				wg.Done()
			}()

			for _, item := range test.Relations {
				t.Logf("pushing relation with source %d and dest %d", item.SourceId, item.TargetId)
				relchan <- item
				res := <-reschan
				if test.ExpectedError != nil {
					t.Logf("entered error branch")
					assert.EqualError(t, res.Error, test.ExpectedError.Error())
				} else {
					assert.NoError(t, res.Error)
				}
			}
			noMoreRels <- struct{}{}
			wg.Wait()

		})
	}
}

func TestController_PushItems(t *testing.T) {

	logger := zap.NewNop()
	var input []drawio.Item
	for i := 1; i < 100; i++ {
		input = append(input, drawio.Item{
			UUID:     "test-push-items",
			EID:      i,
			Class:    drawio.ItemClassResistors,
			SubClass: "resistor_1",
		})
	}

	for i := 70; i < 90; i++ {
		input = append(input, drawio.Item{
			UUID:     "test-push-items",
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
	reschan := make(chan drawio.Item, 100)

	rand.Seed(time.Now().UnixNano())
	perm := rand.Perm(len(input))
	permInput := make([]drawio.Item, len(input))
	for i, v := range perm {
		permInput[v] = input[i]
	}

	tests := []struct {
		Name          string
		ExpectedError error
		items         []drawio.Item
		Reschan       chan drawio.Item
	}{
		{
			"all good",
			nil,
			permInput,
			reschan,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var wg sync.WaitGroup
			noMoreItems := make(chan struct{})
			wg.Add(1)

			go func() {
				ctrlr.PushItems(logger, ichan, reschan, noMoreItems)
				wg.Done()
			}()

			for _, item := range test.items {
				ichan <- item
				t.Logf("pushing item of type %s", item.Class)

			}

			noMoreItems <- struct{}{}
			for res := range reschan {
				if test.ExpectedError != nil {
					t.Logf("entered error branch")
					assert.EqualError(t, res.Error, test.ExpectedError.Error())
				} else {
					assert.NoError(t, res.Error)
				}

			}
		})
	}

}

func Test_channels(t *testing.T) {

	queue := make(chan int, 10)
	done := make(chan struct{}, 1)
	var wg sync.WaitGroup
	wg.Add(1)

	go func(d chan struct{}) {
		defer wg.Done()
	OuterLoop:
		for {
			select {
			case val, ok := <-queue:
				if ok {
					t.Logf("val %d", val)

				} else {
					t.Logf("channel is closed")
				}
			case _, dok := <-d:
				if dok {
					break OuterLoop
				}
			default:

			}
		}
		t.Logf("bye goroutine")
		return
	}(done)

	for i := 0; i < 10; i++ {

		if i == 4 {
			done <- struct{}{}
			//close(queue)
			break
		}

		t.Logf("pushing item")
		queue <- i
		//time.Sleep(1 * time.Nanosecond)
	}

	wg.Wait()
	//close(queue)
}

func TestBuffhan(t *testing.T) {
	bch := make(chan int, 100)
	for i := 1; i < 10; i++ {
		bch <- i
	}

	t.Run("read not full buff", func(t *testing.T) {
		for b := range bch {
			t.Logf("got v=%d", b)
		}
	})
}
