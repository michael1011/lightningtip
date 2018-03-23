package backends

import (
	"context"
	"encoding/hex"
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
)

type LND struct {
	RPCHost      string `long:"rpchost" Description:"Host that the gRPC interface of LND is listening to"`
	CertFile     string `long:"certfile" Description:"TLS certificate for LND gRPC and REST services"`
	MacaroonFile string `long:"macaroonfile" Description:"Admin macaroon file for LND authentication. Set to an empty string for no macaroon"`

	client lnrpc.LightningClient
	ctx    context.Context
}

func (lnd *LND) Connect() error {
	creds, err := credentials.NewClientTLSFromFile(lnd.CertFile, "")

	if err != nil {
		log.Error("Failed to read certificate for LND gRPC")

		return err
	}

	con, err := grpc.Dial(lnd.RPCHost, grpc.WithTransportCredentials(creds))

	if err != nil {
		log.Error("Failed to connect to LND gRPC server")

		return err
	}

	lnd.ctx = context.Background()

	if lnd.MacaroonFile != "" {
		macaroon, err := getMacaroon(lnd.MacaroonFile)

		if macaroon == nil && err != nil {
			log.Error("Failed to read admin macaroon file of LND")
		}

		lnd.ctx = metadata.NewOutgoingContext(lnd.ctx, macaroon)
	}

	lnd.client = lnrpc.NewLightningClient(con)

	return err
}

func getMacaroon(macaroonFile string) (macaroon metadata.MD, err error) {
	data, err := ioutil.ReadFile(macaroonFile)

	if err == nil {
		macaroon = metadata.Pairs("macaroon", hex.EncodeToString(data))
	}

	return macaroon, err
}

func (lnd *LND) GetInvoice(description string, value int64, expiry int64) (invoice string, err error) {
	response, err := lnd.client.AddInvoice(lnd.ctx, &lnrpc.Invoice{
		Memo:   description,
		Value:  value,
		Expiry: expiry,
	})

	if err != nil {
		return "", err
	}

	return response.PaymentRequest, err
}
