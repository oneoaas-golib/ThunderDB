/*
 * Copyright 2018 The ThunderDB Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cpuminer

import (
	"math/big"
	"testing"

	"time"

	"sync"

	"github.com/thunderdb/ThunderDB/crypto/hash"
)

func TestCPUMiner_HashBlock(t *testing.T) {
	miner := NewCPUMiner(make(chan struct{}))
	nonceCh := make(chan Nonce)
	stop := make(chan struct{})
	diffWanted := 20
	data := []byte{
		0x79, 0xa6, 0x1a, 0xdb, 0xc6, 0xe5, 0xa2, 0xe1,
		0x39, 0xd2, 0x71, 0x3a, 0x54, 0x6e, 0xc7, 0xc8,
		0x75, 0x63, 0x2e, 0x75, 0xf1, 0xdf, 0x9c, 0x3f,
		0xa6, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	block := MiningBlock{
		Data:      data,
		NonceChan: nonceCh,
		Stop:      stop,
	}
	var (
		err error
	)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = miner.CalculateBlockNonce(block, *big.NewInt(0), diffWanted)
		wg.Done()
	}()
	nonceFromCh := <-nonceCh
	wg.Wait()

	hash := hash.DoubleHashH(append(data, nonceFromCh.Nonce.Bytes()...))
	if err != nil || nonceFromCh.Difficulty < diffWanted || hash.Difficulty() < diffWanted {
		t.Errorf("CalculateBlockNonce got %v, difficulty %d, nonce %s",
			err, nonceFromCh.Difficulty, nonceFromCh.Nonce.String())
	}
	t.Logf("Difficulty: %d, Hash: %s", nonceFromCh.Difficulty, hash.String())
}

func TestCPUMiner_HashBlock_stop(t *testing.T) {
	minerQuit := make(chan struct{})
	miner := NewCPUMiner(minerQuit)
	nonceCh := make(chan Nonce)
	stop := make(chan struct{})
	diffWanted := 256
	data := []byte{
		0x79, 0xa6,
	}
	block := MiningBlock{
		Data:      data,
		NonceChan: nonceCh,
		Stop:      stop,
	}
	var (
		err error
	)
	go func() {
		err = miner.CalculateBlockNonce(block, *big.NewInt(0), diffWanted)
	}()
	// stop miner
	time.Sleep(2 * time.Second)
	block.Stop <- struct{}{}
	//miner.quit <- struct{}{}

	nonceFromCh := <-block.NonceChan

	//hasha := miner.HashBlock(data, nonceFromCh.Nonce)
	hasha := hash.DoubleHashH(append(data, nonceFromCh.Nonce.Bytes()...))
	if nonceFromCh.Difficulty < 1 || hasha.Difficulty() != nonceFromCh.Difficulty {
		t.Errorf("CalculateBlockNonce got %v, difficulty %d, nonce %s, hash %s",
			err, nonceFromCh.Difficulty, nonceFromCh.Nonce.String(), hasha.String())
	}
	t.Logf("Difficulty: %d, Hash: %s", nonceFromCh.Difficulty, hasha.String())
}

func TestCPUMiner_HashBlock_quit(t *testing.T) {
	minerQuit := make(chan struct{})
	miner := NewCPUMiner(minerQuit)
	nonceCh := make(chan Nonce)
	stop := make(chan struct{})
	diffWanted := 256
	data := []byte{
		0x79, 0xa6,
	}
	block := MiningBlock{
		Data:      data,
		NonceChan: nonceCh,
		Stop:      stop,
	}
	var (
		err error
	)
	go func() {
		err = miner.CalculateBlockNonce(block, *big.NewInt(0), diffWanted)
	}()
	// stop miner
	time.Sleep(2 * time.Second)
	//block.Stop <- struct{}{}
	miner.quit <- struct{}{}

	nonceFromCh := <-block.NonceChan

	//hasha := miner.HashBlock(data, nonceFromCh.Nonce)
	hasha := hash.DoubleHashH(append(data, nonceFromCh.Nonce.Bytes()...))
	if nonceFromCh.Difficulty < 1 || hasha.Difficulty() != nonceFromCh.Difficulty {
		t.Errorf("CalculateBlockNonce got %v, difficulty %d, nonce %s, hash %s",
			err, nonceFromCh.Difficulty, nonceFromCh.Nonce.String(), hasha.String())
	}
	t.Logf("Difficulty: %d, Hash: %s", nonceFromCh.Difficulty, hasha.String())
}
