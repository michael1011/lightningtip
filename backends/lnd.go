package backends

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/donovanhide/eventsource"
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"io"
	"io/ioutil"
)

type LND struct {
	GRPCHost     string `long:"grpchost" Description:"Host of the gRPC interface of LND"`
	CertFile     string `long:"certfile" Description:"TLS certificate for the LND gRPC and REST services"`
	MacaroonFile string `long:"macaroonfile" Description:"Admin macaroon file for authentication. Set to an empty string for no macaroon"`

	client lnrpc.LightningClient
	ctx    context.Context
}

func (lnd *LND) Connect() error {
	creds, err := credentials.NewClientTLSFromFile(lnd.CertFile, "")

	if err != nil {
		log.Error("Failed to read certificate for LND gRPC")

		return err
	}

	con, err := grpc.Dial(lnd.GRPCHost, grpc.WithTransportCredentials(creds))

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

func (lnd *LND) GetInvoice(message string, amount int64, expiry int64) (invoice string, err error) {
	var response *lnrpc.AddInvoiceResponse

	if message != "" {
		response, err = lnd.client.AddInvoice(lnd.ctx, &lnrpc.Invoice{
			Memo:   message,
			Value:  amount,
			Expiry: expiry,
		})

	} else {
		response, err = lnd.client.AddInvoice(lnd.ctx, &lnrpc.Invoice{
			Value:  amount,
			Expiry: expiry,
		})
	}

	if err != nil {
		return "", err
	}

	return response.PaymentRequest, err
}

func (lnd *LND) SubscribeInvoices(callback PublishInvoiceSettled, eventSrv *eventsource.Server) error {
	stream, err := lnd.client.SubscribeInvoices(lnd.ctx, &lnrpc.InvoiceSubscription{})

	if err != nil {
		return err
	}

	wait := make(chan struct{})

	go func() {
		for {
			invoice, streamErr := stream.Recv()

			if streamErr == io.EOF {
				err = errors.New("lost connection to LND gRPC")

				close(wait)

				return
			}

			if streamErr != nil {
				err = streamErr

				close(wait)

				return
			}

			if invoice.Settled {
				callback(invoice.PaymentRequest, eventSrv)
			}

		}

	}()

	<-wait

	return err
}

func getMacaroon(macaroonFile string) (macaroon metadata.MD, err error) {
	data, err := ioutil.ReadFile(macaroonFile)

	if err == nil {
		macaroon = metadata.Pairs("macaroon", hex.EncodeToString(data))
	}

	return macaroon, err
}
