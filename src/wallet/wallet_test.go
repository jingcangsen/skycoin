package wallet

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/util/fee"
)

func TestNewWallet(t *testing.T) {
	type expect struct {
		meta map[string]string
		err  error
	}

	tt := []struct {
		name    string
		wltName string
		ops     []Option
		expect  expect
	}{
		{
			"ok",
			"test.wlt",
			[]Option{},
			expect{
				meta: map[string]string{
					"label":    "",
					"filename": "test.wlt",
					"coin":     "skycoin",
					"type":     "deterministic",
				},
				err: nil,
			},
		},
		{
			"ok with label set",
			"test.wlt",
			[]Option{OptLabel("wallet1")},
			expect{
				meta: map[string]string{
					"label":    "wallet1",
					"filename": "test.wlt",
					"coin":     "skycoin",
					"type":     "deterministic",
				},
				err: nil,
			},
		},
		{
			"ok with label set",
			"test.wlt",
			[]Option{OptLabel("wallet1")},
			expect{
				meta: map[string]string{
					"label":    "wallet1",
					"filename": "test.wlt",
					"coin":     "skycoin",
					"type":     "deterministic",
				},
				err: nil,
			},
		},
		{
			"ok with coin set",
			"test.wlt",
			[]Option{OptLabel("wallet1"), OptCoin("testcoin")},
			expect{
				meta: map[string]string{
					"label":    "wallet1",
					"filename": "test.wlt",
					"coin":     "testcoin",
					"type":     "deterministic",
				},
				err: nil,
			},
		},
		{
			"ok with seed set",
			"test.wlt",
			[]Option{OptLabel("wallet1"), OptSeed("testseed123")},
			expect{
				meta: map[string]string{
					"label":    "wallet1",
					"filename": "test.wlt",
					"coin":     "skycoin",
					"seed":     "testseed123",
					"type":     "deterministic",
				},
				err: nil,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			w, err := NewWallet(tc.wltName, tc.ops...)
			require.Equal(t, tc.expect.err, err)
			if err != nil {
				return
			}
			require.NoError(t, w.Validate())
			for k, v := range tc.expect.meta {
				vv, ok := w.Meta[k]
				require.True(t, ok)
				require.Equal(t, v, vv)
			}
		})
	}
}

func TestLoadWallet(t *testing.T) {
	type expect struct {
		meta map[string]string
		err  error
	}

	tt := []struct {
		name   string
		file   string
		expect expect
	}{
		{
			"ok",
			"./testdata/test1.wlt",
			expect{
				meta: map[string]string{
					"coin":     "sky",
					"filename": "test1.wlt",
					"label":    "test3",
					"lastSeed": "9182b02c0004217ba9a55593f8cf0abecc30d041e094b266dbb5103e1919adaf",
					"seed":     "buddy fossil side modify turtle door label grunt baby worth brush master",
					"tm":       "1503458909",
					"type":     "deterministic",
					"version":  "0.1",
				},
				err: nil,
			},
		},
		{
			"wallet file doesn't exist",
			"not_exist_file.wlt",
			expect{
				meta: map[string]string{},
				err:  fmt.Errorf("load wallet file failed, wallet not_exist_file.wlt doesn't exist"),
			},
		},
		{
			"invalid wallet: no type",
			"./testdata/invalid_wallets/no_type.wlt",
			expect{
				meta: map[string]string{},
				err:  fmt.Errorf("invalid wallet no_type.wlt: type field not set"),
			},
		},
		{
			"invalid wallet: invalid type",
			"./testdata/invalid_wallets/err_type.wlt",
			expect{
				meta: map[string]string{},
				err:  fmt.Errorf("invalid wallet err_type.wlt: wallet type invalid"),
			},
		},
		{
			"invalid wallet: no coin",
			"./testdata/invalid_wallets/no_coin.wlt",
			expect{
				meta: map[string]string{},
				err:  fmt.Errorf("invalid wallet no_coin.wlt: coin field not set"),
			},
		},
		{
			"invalid wallet: no seed",
			"./testdata/invalid_wallets/no_seed.wlt",
			expect{
				meta: map[string]string{},
				err:  fmt.Errorf("invalid wallet no_seed.wlt: seed field not set"),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			w := Wallet{}
			err := w.Load(tc.file)
			require.Equal(t, tc.expect.err, err)
			if err != nil {
				return
			}

			for k, v := range tc.expect.meta {
				vv := w.Meta[k]
				require.Equal(t, v, vv)
			}
		})
	}
}

func TestWalletGetEntry(t *testing.T) {
	tt := []struct {
		name    string
		wltFile string
		address string
		find    bool
	}{
		{
			"ok",
			"./testdata/test1.wlt",
			"JUdRuTiqD1mGcw358twMg3VPpXpzbkdRvJ",
			true,
		},
		{
			"entry not exist",
			"./testdata/test1.wlt",
			"2ULfxDUuenUY5V4Pr8whmoAwFdUseXNyjXC",
			false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			w := Wallet{}
			require.NoError(t, w.Load(tc.wltFile))
			a, err := cipher.DecodeBase58Address(tc.address)
			require.NoError(t, err)
			e, ok := w.GetEntry(a)
			require.Equal(t, tc.find, ok)
			if ok {
				require.Equal(t, tc.address, e.Address.String())
			}
		})
	}
}

func TestWalletAddEntry(t *testing.T) {
	test1SecKey, err := cipher.SecKeyFromHex("1fc5396e91e60b9fc613d004ea5bd2ccea17053a12127301b3857ead76fdb93e")
	require.NoError(t, err)

	_, s := cipher.GenerateKeyPair()
	seckeys := []cipher.SecKey{
		test1SecKey,
		s,
	}

	tt := []struct {
		name    string
		wltFile string
		secKey  cipher.SecKey
		err     error
	}{
		{
			"ok",
			"./testdata/test1.wlt",
			seckeys[1],
			nil,
		},
		{
			"dup entry",
			"./testdata/test1.wlt",
			seckeys[0],
			errors.New("duplicate address entry"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			w := Wallet{}
			require.NoError(t, w.Load(tc.wltFile))
			a := cipher.AddressFromSecKey(tc.secKey)
			p := cipher.PubKeyFromSecKey(tc.secKey)
			require.Equal(t, tc.err, w.AddEntry(Entry{
				Address: a,
				Public:  p,
				Secret:  s,
			}))
		})
	}
}

func TestWalletDistributeSpendHours(t *testing.T) {
	cases := []struct {
		name              string
		inputHours        uint64
		nAddrs            uint64
		haveChange        bool
		expectChangeHours uint64
		expectAddrHours   []uint64
	}{
		{
			name:            "no input hours, one addr, no change",
			inputHours:      0,
			nAddrs:          1,
			haveChange:      false,
			expectAddrHours: []uint64{0},
		},
		{
			name:            "no input hours, two addrs, no change",
			inputHours:      0,
			nAddrs:          2,
			haveChange:      false,
			expectAddrHours: []uint64{0, 0},
		},
		{
			name:            "no input hours, one addr, change",
			inputHours:      0,
			nAddrs:          1,
			haveChange:      true,
			expectAddrHours: []uint64{0},
		},
		{
			name:            "one input hour, one addr, no change",
			inputHours:      1,
			nAddrs:          1,
			haveChange:      false,
			expectAddrHours: []uint64{0},
		},
		{
			name:            "two input hours, one addr, no change",
			inputHours:      2,
			nAddrs:          1,
			haveChange:      false,
			expectAddrHours: []uint64{1},
		},
		{
			name:              "two input hours, one addr, change",
			inputHours:        2,
			nAddrs:            1,
			haveChange:        true,
			expectChangeHours: 1,
			expectAddrHours:   []uint64{0},
		},
		{
			name:              "three input hours, one addr, change",
			inputHours:        3,
			nAddrs:            1,
			haveChange:        true,
			expectChangeHours: 1,
			expectAddrHours:   []uint64{0},
		},
		{
			name:            "three input hours, one addr, no change",
			inputHours:      3,
			nAddrs:          1,
			haveChange:      false,
			expectAddrHours: []uint64{1},
		},
		{
			name:            "three input hours, two addrs, no change",
			inputHours:      3,
			nAddrs:          2,
			haveChange:      false,
			expectAddrHours: []uint64{1, 0},
		},
		{
			name:            "four input hours, one addr, no change",
			inputHours:      4,
			nAddrs:          1,
			haveChange:      false,
			expectAddrHours: []uint64{2},
		},
		{
			name:              "four input hours, one addr, change",
			inputHours:        4,
			nAddrs:            1,
			haveChange:        true,
			expectChangeHours: 1,
			expectAddrHours:   []uint64{1},
		},
		{
			name:              "four input hours, two addr, change",
			inputHours:        4,
			nAddrs:            2,
			haveChange:        true,
			expectChangeHours: 1,
			expectAddrHours:   []uint64{1, 0},
		},
		{
			name:              "30 (divided by 2, odd number) input hours, two addr, change",
			inputHours:        30,
			nAddrs:            2,
			haveChange:        true,
			expectChangeHours: 8,
			expectAddrHours:   []uint64{4, 3},
		},
		{
			name:              "33 (odd number) input hours, two addr, change",
			inputHours:        33,
			nAddrs:            2,
			haveChange:        true,
			expectChangeHours: 8,
			expectAddrHours:   []uint64{4, 4},
		},
		{
			name:              "33 (odd number) input hours, three addr, change",
			inputHours:        33,
			nAddrs:            3,
			haveChange:        true,
			expectChangeHours: 8,
			expectAddrHours:   []uint64{3, 3, 2},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			changeHours, addrHours, totalHours := DistributeSpendHours(tc.inputHours, tc.nAddrs, tc.haveChange)
			require.Equal(t, tc.expectChangeHours, changeHours)
			require.Equal(t, tc.expectAddrHours, addrHours)
			require.Equal(t, tc.nAddrs, uint64(len(addrHours)))

			outputHours := changeHours
			for _, h := range addrHours {
				outputHours += h
			}
			require.True(t, tc.inputHours >= outputHours)
			require.Equal(t, outputHours, totalHours)

			if tc.inputHours != 0 {
				err := fee.VerifyTransactionFeeForHours(outputHours, tc.inputHours-outputHours)
				require.NoError(t, err)
			}
		})
	}

	// Tests over range of values
	for inputHours := uint64(0); inputHours <= 1e3; inputHours++ {
		for nAddrs := uint64(1); nAddrs < 16; nAddrs++ {
			for _, haveChange := range []bool{true, false} {
				name := fmt.Sprintf("inputHours=%d nAddrs=%d haveChange=%v", inputHours, nAddrs, haveChange)
				t.Run(name, func(t *testing.T) {
					changeHours, addrHours, totalHours := DistributeSpendHours(inputHours, nAddrs, haveChange)
					require.Equal(t, nAddrs, uint64(len(addrHours)))

					var sumAddrHours uint64
					for _, h := range addrHours {
						sumAddrHours += h
					}

					if haveChange {
						expectedChangeHours := (inputHours / fee.BurnFactor) / 2
						require.True(t, changeHours == expectedChangeHours || changeHours == expectedChangeHours+1)
						require.Equal(t, expectedChangeHours, sumAddrHours)
					} else {
						require.Equal(t, uint64(0), changeHours)
						require.Equal(t, inputHours/fee.BurnFactor, sumAddrHours)
					}

					outputHours := sumAddrHours + changeHours
					require.True(t, inputHours >= outputHours)
					require.Equal(t, outputHours, totalHours)

					if inputHours != 0 {
						err := fee.VerifyTransactionFeeForHours(outputHours, inputHours-outputHours)
						require.NoError(t, err)
					}

					// addrHours at the beginning and end of the array should not differ by more than one
					max := addrHours[0]
					min := addrHours[len(addrHours)-1]
					require.True(t, max-min <= 1)
				})
			}
		}
	}
}
