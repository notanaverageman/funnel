//go:build !disableelasticsearch
// +build !disableelasticsearch

package outputs

// This is the elasticsearch output writer
import (
	"fmt"

	"github.com/agnivade/funnel"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
)

// Registering the constructor function
func init() {
	funnel.RegisterNewWriter("elasticsearch", newElasticSearchOutput)
}

// ESLogger is a wrapper over logrus to satisfy the logger interface of
// elasticsearch client
type ESLogger struct {
}

// Printf calls the Err() of the logrus instead
func (el *ESLogger) Printf(format string, v ...interface{}) {
	logrus.Error(fmt.Sprintf(format, v...))
}

func newElasticSearchOutput(v *viper.Viper) (funnel.OutputWriter, error) {
	// Creating elastic client
	c, err := elastic.NewClient(
		elastic.SetURL(v.GetStringSlice("target.nodes")...),
		elastic.SetGzip(true),
		elastic.SetErrorLog(&ESLogger{}),
		elastic.SetBasicAuth(v.GetString("target.username"), v.GetString("target.password")))

	if err != nil {
		return nil, err
	}

	// Creating the struct
	e := &elasticOutput{
		bulkSvc:   c.Bulk(),
		index:     v.GetString("target.index"),
		indexType: v.GetString("target.type"),
	}
	return e, nil
}

type elasticOutput struct {
	bulkSvc   *elastic.BulkService
	index     string
	indexType string
}

// Implmenting the OutputWriter interface

func (e *elasticOutput) Write(p []byte) (n int, err error) {
	// Adding a document to the bulk request
	bulkReq := elastic.NewBulkIndexRequest().
		Doc(string(p)).
		Index(e.index).
		Type(e.indexType)
	e.bulkSvc.Add(bulkReq)
	return len(p), nil
}

func (e *elasticOutput) Flush() error {
	// Sends all bulked request to elasticsearch
	_, err := e.bulkSvc.Do(context.TODO())
	return err
}

func (e *elasticOutput) Close() error {
	return nil
}
