package ursa

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCredentialValues_AddValue(t *testing.T) {
	type args struct {
		name string
		raw  interface{}
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{name: "address2", args: args{name: "address2", raw: "101 Wilson Lane"}, expected: "68086943237164982734333428280784300550565381723532936263016368251445461241953"},
		{name: "zip", args: args{name: "zip", raw: "87121"}, expected: "87121"},
		{name: "city", args: args{name: "city", raw: "SLC"}, expected: "101327353979588246869873249766058188995681113722618593621043638294296500696424"},
		{name: "address1", args: args{name: "address1", raw: "101 Tela Lane"}, expected: "63690509275174663089934667471948380740244018358024875547775652380902762701972"},
		{name: "state", args: args{name: "state", raw: "UT"}, expected: "93856629670657830351991220989031130499313559332549427637940645777813964461231"},
		{name: "Empty", args: args{name: "Empty", raw: ""}, expected: "102987336249554097029535212322581322789799900648198034993379397001115665086549"},
		{name: "Null", args: args{name: "Null", raw: nil}, expected: "99769404535520360775991420569103450442789945655240760487761322098828903685777"},
		{name: "bool True", args: args{name: "bool True", raw: true}, expected: "1"},
		{name: "bool False", args: args{name: "bool False", raw: false}, expected: "0"},
		{name: "str True", args: args{name: "str True", raw: "True"}, expected: "27471875274925838976481193902417661171675582237244292940724984695988062543640"},
		{name: "str False", args: args{name: "str False", raw: "False"}, expected: "43710460381310391454089928988014746602980337898724813422905404670995938820350"},
		{name: "max i32", args: args{name: "max i32", raw: 2147483647}, expected: "2147483647"},
		{name: "max i32 + 1", args: args{name: "max i32 + 1", raw: 2147483648}, expected: "26221484005389514539852548961319751347124425277437769688639924217837557266135"},
		{name: "min i32", args: args{name: "min i32", raw: -2147483648}, expected: "-2147483648"},
		{name: "min i32 - 1", args: args{name: "min i32 - 1", raw: -2147483649}, expected: "68956915425095939579909400566452872085353864667122112803508671228696852865689"},
		{name: "float 0.0", args: args{name: "float 0.0", raw: 0.0}, expected: "62838607218564353630028473473939957328943626306458686867332534889076311281879"},
		{name: "str 0.0", args: args{name: "str 0.0", raw: "0.0"}, expected: "62838607218564353630028473473939957328943626306458686867332534889076311281879"},
		//{name: "chr 0", args: args{name: "chr 0", raw: "chr(0)"}, expected: "49846369543417741186729467304575255505141344055555831574636310663216789168157"},
		//{name: "chr 1", args: args{name: "chr 1", raw: "chr(1)"}, expected: "34356466678672179216206944866734405838331831190171667647615530531663699592602"},
		//{name: "chr 2", args: args{name: "chr 2", raw: "chr(2)"}, expected: "99398763056634537812744552006896172984671876672520535998211840060697129507206"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewValues()
			r.AddValue(tt.args.name, tt.args.raw)
			result, err := json.MarshalIndent(r, " ", " ")
			require.NoError(t, err)
			m := map[string]interface{}{}
			err = json.Unmarshal(result, &m)
			require.NoError(t, err)
			vals, ok := m[tt.args.name].(map[string]interface{})
			require.True(t, ok)

			require.EqualValues(t, tt.args.raw, vals["raw"], tt.name)
			require.Equal(t, tt.expected, vals["encoded"].(string), tt.name)
		})
	}
}
