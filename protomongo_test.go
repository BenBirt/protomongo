package protomongo_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/BenBirt/protomongo"
	pb "github.com/BenBirt/protomongo/example"
	mongodb "github.com/BenBirt/protomongo/mongodb/testing"
	"github.com/golang/protobuf/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	simpleMessage = &pb.SimpleMessage{
		StringField: "foo",
		Int32Field:  32525,
		Int64Field:  1531541553141312315,
		FloatField:  21541.3242,
		DoubleField: 21535215136361617136.543858,
		BoolField:   true,
		EnumField:   pb.Enum_VAL_2,
	}
	tests = []struct {
		name          string
		pb            proto.Message
		equivalentPbs []proto.Message
	}{
		{
			name: "simple message",
			pb:   simpleMessage,
			equivalentPbs: []proto.Message{
				&pb.RepeatedFieldMessage{
					StringField: []string{"foo"},
					Int32Field:  []int32{32525},
					Int64Field:  []int64{1531541553141312315},
					FloatField:  []float32{21541.3242},
					DoubleField: []float64{21535215136361617136.543858},
					BoolField:   []bool{true},
					EnumField:   []pb.Enum{pb.Enum_VAL_2},
				},
			},
		},
		{
			name: "message with repeated fields",
			pb: &pb.RepeatedFieldMessage{
				StringField: []string{"foo", "bar"},
				Int32Field:  []int32{32525, 1958, 435},
				Int64Field:  []int64{1531541553141312315, 13512516266},
				FloatField:  []float32{21541.3242, 634214.2233, 3435.322},
				DoubleField: []float64{21535215136361617136.543858, 213143343.76767},
				BoolField:   []bool{true, false, true, true},
				EnumField:   []pb.Enum{pb.Enum_VAL_2, pb.Enum_VAL_1},
			},
			equivalentPbs: []proto.Message{
				&pb.SimpleMessage{
					StringField: "bar",
					Int32Field:  435,
					Int64Field:  13512516266,
					FloatField:  3435.322,
					DoubleField: 213143343.76767,
					BoolField:   true,
					EnumField:   pb.Enum_VAL_1,
				},
			},
		},
		{
			name: "message with submessage",
			pb: &pb.MessageWithSubMessage{
				StringField: "baz",
				SimpleMessage: &pb.SimpleMessage{
					StringField: "foo",
					Int32Field:  32525,
					Int64Field:  1531541553141312315,
					FloatField:  21541.3242,
					DoubleField: 21535215136361617136.543858,
					BoolField:   true,
					EnumField:   pb.Enum_VAL_2,
				},
			},
			equivalentPbs: []proto.Message{
				&pb.MessageWithRepeatedSubMessage{
					StringField: "baz",
					SimpleMessage: []*pb.SimpleMessage{
						{
							StringField: "foo",
							Int32Field:  32525,
							Int64Field:  1531541553141312315,
							FloatField:  21541.3242,
							DoubleField: 21535215136361617136.543858,
							BoolField:   true,
							EnumField:   pb.Enum_VAL_2,
						},
					},
				},
			},
		},
		{
			name: "message with repeated submessage",
			pb: &pb.MessageWithRepeatedSubMessage{
				StringField: "baz",
				SimpleMessage: []*pb.SimpleMessage{
					{
						StringField: "foo",
						Int32Field:  32525,
						Int64Field:  1531541553141312315,
						FloatField:  21541.3242,
						DoubleField: 21535215136361617136.543858,
						BoolField:   true,
						EnumField:   pb.Enum_VAL_2,
					},
					{
						StringField: "qux",
						Int32Field:  22,
						BoolField:   false,
					},
				},
			},
			equivalentPbs: []proto.Message{
				&pb.MessageWithSubMessage{
					StringField: "baz",
					SimpleMessage: &pb.SimpleMessage{
						StringField: "qux",
						Int32Field:  22,
						Int64Field:  1531541553141312315,
						FloatField:  21541.3242,
						DoubleField: 21535215136361617136.543858,
						// It might be expected that because the last element of the 'SimpleMessage' slice in 'pb' explicitly sets 'BoolField' to false,
						// this field should also be false, because the elements of the 'SimpleMessage' slice should be merged in order.
						// However, by the rules of proto3, default field values are never serialized. Thus when the second element
						// of the 'SimpleMessage' slice is deserialized, that deserialized value contains no value for 'BoolField', and thus
						// this field retains the value that was set in the first element of that slice.
						BoolField: true,
						EnumField: pb.Enum_VAL_2,
					},
				},
			},
		},
		{
			name: "message with oneof",
			pb: &pb.MessageWithOneof{
				StringField: "baz",
				OneofField:  &pb.MessageWithOneof_Int32OneofField{3132},
			},
			equivalentPbs: []proto.Message{},
		},
	}
)

func TestAgainstRealDatabase(t *testing.T) {
	db, stopMongo := startMongoDatabase(t)
	defer stopMongo()
	coll := db.Collection("test_collection")
	if _, err := coll.InsertOne(context.Background(), simpleMessage); err != nil {
		t.Errorf("coll.InsertOne(%v) error = %v, want nil", simpleMessage, err)
	}
	count, err := coll.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		t.Errorf("coll.CountDocuments() error = %v, want nil", err)
	}
	if count != 1 {
		t.Errorf("coll.CountDocuments() = %v; want 1", count)
	}
	var found *pb.SimpleMessage
	if err := coll.FindOne(context.Background(), bson.D{}).Decode(&found); err != nil {
		t.Errorf("coll.FindOne().Decode() error = %v, want nil", err)
	}
	if !proto.Equal(simpleMessage, found) {
		t.Errorf("proto.Equal(%v, %v) = false, want true", simpleMessage, found)
	}
}

// TODO: Add a testcase looking up protos by fields/nested fields.

func TestMarshalUnmarshal(t *testing.T) {
	rb := bson.NewRegistryBuilder()
	rb.RegisterCodec(reflect.TypeOf((*proto.Message)(nil)).Elem(), protomongo.NewProtobufCodec())
	reg := rb.Build()

	for _, testCase := range tests {
		b, err := bson.MarshalWithRegistry(reg, testCase.pb)
		if err != nil {
			t.Errorf("bson.MarshalWithRegistry(%v) error = %v, want nil", testCase.pb, err)
		}

		for _, equivalentPb := range append(testCase.equivalentPbs, testCase.pb) {
			out := reflect.New(reflect.TypeOf(equivalentPb).Elem()).Interface().(proto.Message)
			if err = bson.UnmarshalWithRegistry(reg, b, &out); err != nil {
				t.Errorf("bson.UnmarshalWithRegistry(%v) error = %v, want nil", b, err)
			}
			if !proto.Equal(equivalentPb, out) {
				t.Errorf("proto.Equal(%v, %v) = false, want true", equivalentPb, out)
			}
		}
	}
}

func TestMarshalUnmarshalWithPointers(t *testing.T) {
	rb := bson.NewRegistryBuilder()
	rb.RegisterCodec(reflect.TypeOf((*proto.Message)(nil)).Elem(), protomongo.NewProtobufCodec())
	reg := rb.Build()

	for _, testCase := range tests {
		b, err := bson.MarshalWithRegistry(reg, testCase.pb)
		if err != nil {
			t.Errorf("bson.MarshalWithRegistry(%v) error = %v, want nil", testCase.pb, err)
		}

		for _, equivalentPb := range append(testCase.equivalentPbs, testCase.pb) {
			out := reflect.New(reflect.TypeOf(equivalentPb).Elem()).Interface().(proto.Message)
			if err = bson.UnmarshalWithRegistry(reg, b, &out); err != nil {
				t.Errorf("bson.UnmarshalWithRegistry(%v) error = %v, want nil", b, err)
			}
			if !proto.Equal(equivalentPb, out) {
				t.Errorf("proto.Equal(%v, %v) = false, want true", equivalentPb, out)
			}
		}
	}
}

func startMongoDatabase(t *testing.T) (*mongo.Database, func()) {
	mongod := &mongodb.Mongod{}
	if err := mongod.Start(); err != nil {
		t.Fatal(err)
	}
	m, err := mongod.GetClient()
	if err != nil {
		t.Fatal(err)
	}
	d := m.Database("test_db")
	return d, func() { mongod.Stop() }
}
