package internal

import (
	"reflect"
	"testing"
)

func Test_Pot(t *testing.T) {
	bets := []uint32{60, 80, 90, 0, 40, 160, 0, 0, 200}
	res := calcPot(bets)
	//[{180 [1 2 3]} {40 [2 3]} {10 [3]}]
	t.Log(res)

	bets = []uint32{60, 60, 60, 0, 0, 0, 0, 0, 0}
	res = calcPot(bets)
	//[{180 [1 2 3]}]
	t.Log(res)

	a := []byte{}

	a = nil
	t.Log(len(a))

}

func Test_calcPot(t *testing.T) {
	type args struct {
		bets []uint32
	}
	tests := []struct {
		name     string
		args     args
		wantPots []handPot
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPots := calcPot(tt.args.bets); !reflect.DeepEqual(gotPots, tt.wantPots) {
				t.Errorf("calcPot() = %v, want %v", gotPots, tt.wantPots)
			}
		})
	}
}
