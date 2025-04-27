package main

import (
	"github.com/alexedwards/scs/v2"
	"github.com/dys2p/bitpay"
	"github.com/dys2p/btcpay"
	"github.com/dys2p/digitalgoods/userdb"
	"github.com/dys2p/eco/lang"
	"github.com/dys2p/ordersystem"
)

type Server struct {
	BitpayClient *bitpay.Client
	BtcPayStore  btcpay.Store
	DB           *ordersystem.DB
	Langs        lang.Languages
	Sessions     *scs.SessionManager
	Users        userdb.Authenticator
}
