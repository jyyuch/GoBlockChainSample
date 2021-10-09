package proxy

import (
	"myModule/model"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fetchBlockByNumber(t *testing.T) {
	assert := assert.New(t)

	type args struct {
		blockStart uint64
		blockEnd   uint64
		inOut      []*model.BlockBase
		wg         *sync.WaitGroup
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test fetch genesis block",
			args: args{
				blockStart: 0,
				blockEnd:   0,
				inOut:      make([]*model.BlockBase, 1),
				wg:         &sync.WaitGroup{},
			},
		},
		{
			name: "test fetch first 2 block",
			args: args{
				blockStart: 0,
				blockEnd:   1,
				inOut:      make([]*model.BlockBase, 1-0+1),
				wg:         &sync.WaitGroup{},
			},
		},
		{
			name: "test fetch first 21 block",
			args: args{
				blockStart: 1,
				blockEnd:   21,
				inOut:      make([]*model.BlockBase, 21-1+1),
				wg:         &sync.WaitGroup{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetchHeaderByNumber(tt.args.blockStart, tt.args.blockEnd, tt.args.inOut, tt.args.wg)

			for i, v := range tt.args.inOut {
				assert.NotNil(v)
				if i > 0 {
					assert.Equal(v.ParentHash, tt.args.inOut[i-1].Hash)
				}
			}
		})
	}
}
