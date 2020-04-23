package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/crypto"
	"github.com/filecoin-project/specs-actors/actors/runtime/exitcode"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-filestore"
	"github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
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

	addExample(bitfield.NewFromSet([]uint64{5}))
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
	addExample(abi.SectorNumber(9))
	addExample(abi.SectorSize(32 * 1024 * 1024 * 1024))
	addExample(api.MpoolChange(0))
	addExample(network.Connected)
	addExample(dtypes.NetworkName("lotus"))
	addExample(api.SyncStateStage(1))
	addExample(build.APIVersion)
	addExample(api.PCHInbound)
	addExample(time.Minute)
	addExample(&types.ExecutionResult{
		Msg:    exampleValue(reflect.TypeOf(&types.Message{})).(*types.Message),
		MsgRct: exampleValue(reflect.TypeOf(&types.MessageReceipt{})).(*types.MessageReceipt),
	})
	addExample(map[string]types.Actor{
		"t01236": exampleValue(reflect.TypeOf(types.Actor{})).(types.Actor),
	})
	addExample(map[string]api.MarketDeal{
		"t026363": exampleValue(reflect.TypeOf(api.MarketDeal{})).(api.MarketDeal),
	})
	addExample(map[string]api.MarketBalance{
		"t026363": exampleValue(reflect.TypeOf(api.MarketBalance{})).(api.MarketBalance),
	})

	maddr, err := multiaddr.NewMultiaddr("/ip4/52.36.61.156/tcp/1347/p2p/12D3KooWFETiESTf1v4PGUvtnxMAcEFMzLZbJGg4tjWfGEimYior")
	if err != nil {
		panic(err)
	}

	// because reflect.TypeOf(maddr) returns the concrete type...
	ExampleValues[reflect.TypeOf(struct{ A multiaddr.Multiaddr }{}).Field(0).Type] = maddr

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
		ExampleValues[t] = v
		return v
	case reflect.Array:
		out := reflect.New(t).Elem()
		for i := 0; i < t.Len(); i++ {
			out.Index(i).Set(reflect.ValueOf(exampleValue(t.Elem())))
		}
		return out.Interface()

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

type Visitor struct {
	Methods map[string]ast.Node
}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	st, ok := node.(*ast.TypeSpec)
	if !ok {
		return v
	}

	if st.Name.Name != "FullNode" {
		return nil
	}

	iface := st.Type.(*ast.InterfaceType)
	for _, m := range iface.Methods.List {
		if len(m.Names) > 0 {
			v.Methods[m.Names[0].Name] = m
		}
	}

	return v
}

func parseApiASTInfo() map[string]string {

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "./api", nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		fmt.Println("parse error: ", err)
	}

	ap := pkgs["api"]

	f := ap.Files["api/api_full.go"]

	cmap := ast.NewCommentMap(fset, f, f.Comments)

	v := &Visitor{make(map[string]ast.Node)}
	ast.Walk(v, pkgs["api"])

	out := make(map[string]string)
	for mn, node := range v.Methods {
		cs := cmap.Filter(node).Comments()
		if len(cs) == 0 {
			out[mn] = "NO COMMENTS"
		} else {
			out[mn] = cs[len(cs)-1].Text()
		}
	}
	return out
}

func main() {
	//b, _ := json.Marshal(exampleValue(reflect.TypeOf(&types.BlockHeader{})))
	//fmt.Println(string(b))
	//return

	comments := parseApiASTInfo()

	var api struct{ api.FullNode }
	t := reflect.TypeOf(api)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)

		fmt.Printf("## `%s`\n", m.Name)

		fmt.Println(comments[m.Name])
		fmt.Println()

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

		fmt.Printf("Inputs: `%s`\n\n", string(v))

		outv := exampleValue(ft.Out(0))

		ov, err := json.Marshal(outv)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Response: `%s`\n\n", string(ov))
		fmt.Println()
	}
}
