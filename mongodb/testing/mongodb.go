package mongodb

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strconv"

	"github.com/BenBirt/protomongo"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/phayes/freeport"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	rootDir = os.Getenv("TEST_TMPDIR") + "/mongodb"
)

func init() {
	if err := os.Mkdir(rootDir, 0777); err != nil && !os.IsExist(err) {
		panic(err)
	}
}

type Mongod struct {
	port int
	cmd  *exec.Cmd
}

func (m *Mongod) Start() error {
	instanceDir := rootDir + "/" + uuid.New().String()
	dbDir := instanceDir + "/db"

	for _, dir := range []string{instanceDir, dbDir} {
		if err := os.Mkdir(dir, 0777); err != nil {
			panic(err)
		}
	}

	port, err := freeport.GetFreePort()
	if err != nil {
		return err
	}
	m.port = port
	m.cmd = exec.Command(
		"external/mongodb/bin/mongo",
		"--port", strconv.Itoa(m.port),
		"--dbpath", dbDir,
	)
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr
	return m.cmd.Start()
}

func (m *Mongod) GetClient() (*mongo.Client, error) {
	rb := bson.NewRegistryBuilder()
	rb.RegisterCodec(reflect.TypeOf((*proto.Message)(nil)).Elem(), protomongo.NewProtobufCodec())
	reg := rb.Build()
	return mongo.Connect(context.Background(), options.Client().ApplyURI(fmt.Sprintf("mongodb://localhost:%v", m.port)).SetRegistry(reg))
}

func (m *Mongod) Stop() {
	m.cmd.Process.Kill()
}
