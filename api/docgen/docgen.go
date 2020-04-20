package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/crypto"
	"github.com/filecoin-project/specs-actors/actors/runtime/exitcode"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-filestore"
	peer "github.com/libp2p/go-libp2p-peer"
)

var ExampleValues = map[reflect.Type]interface{}{
	reflect.TypeOf(api.Permission("")): api.Permission("write"),
	reflect.TypeOf(""):                 "string value",
	reflect.TypeOf(uint64(42)):         uint64(42),
	reflect.TypeOf(byte(7)):            byte(7),
	reflect.TypeOf([]byte{}):           []byte("byte array"),
}

func addExample(v interface{}) {
	ExampleValues[reflect.TypeOf(v)] = v
}

func init() {
	c, err := cid.Decode("bafy2bzacea3wsdh6y3a36tb3skempjoxqpuyompjbmfeyf34fi3uy6uue42v4")
	if err != nil {
		panic(err)
	}

	ExampleValues[reflect.TypeOf(c)] = c

	c2, err := cid.Decode("bafy2bzacebp3shtrn43k7g3unredz7fxn4gj533d3o43tqn2p2ipxxhrvchve")
	if err != nil {
		panic(err)
	}

	tsk := types.NewTipSetKey(c, c2)

	ExampleValues[reflect.TypeOf(tsk)] = tsk

	addr, err := address.NewIDAddress(1234)
	if err != nil {
		panic(err)
	}

	ExampleValues[reflect.TypeOf(addr)] = addr

	pid, err := peer.IDB58Decode("12D3KooWGzxzKZYveHXtpG6AsrUJBcWxHBFS2HsEoGTxrMLvKXtf")
	if err != nil {
		panic(err)
	}
	addExample(pid)

	addExample(abi.RegisteredProof_StackedDRG32GiBPoSt)
	addExample(abi.ChainEpoch(10101))
	addExample(crypto.SigTypeBLS)
	addExample(int64(9))
	addExample(abi.MethodNum(1))
	addExample(exitcode.ExitCode(0))
	addExample(crypto.DomainSeparationTag_ElectionPoStChallengeSeed)
	addExample(true)
	addExample(abi.UnpaddedPieceSize(1024))
	addExample(abi.UnpaddedPieceSize(1024).Padded())
	addExample(abi.DealID(5432))
	addExample(filestore.StatusFileChanged)
}

func exampleValue(t reflect.Type) interface{} {
	v, ok := ExampleValues[t]
	if ok {
		return v
	}

	switch t.Kind() {
	case reflect.Slice:
		out := reflect.New(t).Elem()
		reflect.Append(out, reflect.ValueOf(exampleValue(t.Elem())))
		return out.Interface()
	case reflect.Chan:
		return exampleValue(t.Elem())
	case reflect.Struct:
		es := exampleStruct(t)
		v := reflect.ValueOf(es).Elem().Interface()
		//ExampleValues[t] = v
		return v
	case reflect.Ptr:
		if t.Elem().Kind() == reflect.Struct {
			es := exampleStruct(t.Elem())
			//ExampleValues[t] = es
			return es
		}
	case reflect.Interface:
		return struct{}{}
	}

	panic(fmt.Sprintf("No example value for type: %s", t))
}

func exampleStruct(t reflect.Type) interface{} {
	ns := reflect.New(t)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if strings.Title(f.Name) == f.Name {
			ns.Elem().Field(i).Set(reflect.ValueOf(exampleValue(f.Type)))
		}
	}

	return ns.Interface()
}

func main() {
	//b, _ := json.Marshal(exampleValue(reflect.TypeOf(&types.BlockHeader{})))
	//fmt.Println(string(b))
	//return

	var api struct{ api.FullNode }
	t := reflect.TypeOf(api)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)

		fmt.Printf("## `%s`\n", m.Name)

		var args []interface{}
		ft := m.Func.Type()
		for j := 2; j < ft.NumIn(); j++ {
			inp := ft.In(j)
			args = append(args, exampleValue(inp))
		}

		v, err := json.Marshal(args)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Inputs: `%s`\n", string(v))

		outv := exampleValue(ft.Out(0))

		ov, err := json.Marshal(outv)
		if err != nil {
			panic(err)
		}

		fmt.Println(ft.Out(0))
		fmt.Printf("Response: `%s`\n", string(ov))
		fmt.Println()
	}
}
