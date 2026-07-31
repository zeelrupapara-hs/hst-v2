package fixquotes

import (
	"fmt"
	"time"

	"hstquote/internal/fixconfig"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	fix43mdr "github.com/quickfixgo/fix43/marketdatarequest"
	fix44mdr "github.com/quickfixgo/fix44/marketdatarequest"
	"github.com/quickfixgo/quickfix"
)

func (c *Connector) sendMarketRequest(symbol string, session quickfix.SessionID) error {
	if c.cfg.Dialect == fixconfig.Dialect43 {
		return c.sendMarketRequest43(symbol, session)
	}
	return c.sendMarketRequest44(symbol, session)
}

func (c *Connector) sendMarketRequest44(symbol string, session quickfix.SessionID) error {
	request := fix44mdr.New(
		field.NewMDReqID(fmt.Sprintf("%s-%d", symbol, time.Now().Unix())),
		field.NewSubscriptionRequestType(enum.SubscriptionRequestType_SNAPSHOT_PLUS_UPDATES),
		field.NewMarketDepth(c.cfg.MarketDepth),
	)
	request.SetMDUpdateType(c.mdUpdateType)

	symbols := fix44mdr.NewNoRelatedSymRepeatingGroup()
	symbols.Add().SetSymbol(symbol)
	request.SetNoRelatedSym(symbols)

	entryTypes := fix44mdr.NewNoMDEntryTypesRepeatingGroup()
	entryTypes.Add().SetMDEntryType(enum.MDEntryType_BID)
	entryTypes.Add().SetMDEntryType(enum.MDEntryType_OFFER)
	request.SetNoMDEntryTypes(entryTypes)

	request.Header.Set(field.NewSenderCompID(session.SenderCompID))
	request.Header.Set(field.NewTargetCompID(session.TargetCompID))

	return quickfix.SendToTarget(request.ToMessage(), session)
}

func (c *Connector) sendMarketRequest43(symbol string, session quickfix.SessionID) error {
	request := fix43mdr.New(
		field.NewMDReqID(fmt.Sprintf("%s-%d", symbol, time.Now().Unix())),
		field.NewSubscriptionRequestType(enum.SubscriptionRequestType_SNAPSHOT_PLUS_UPDATES),
		field.NewMarketDepth(c.cfg.MarketDepth),
	)
	request.SetMDUpdateType(c.mdUpdateType)

	symbols := fix43mdr.NewNoRelatedSymRepeatingGroup()
	symbols.Add().SetSymbol(symbol)
	request.SetNoRelatedSym(symbols)

	entryTypes := fix43mdr.NewNoMDEntryTypesRepeatingGroup()
	entryTypes.Add().SetMDEntryType(enum.MDEntryType_BID)
	entryTypes.Add().SetMDEntryType(enum.MDEntryType_OFFER)
	request.SetNoMDEntryTypes(entryTypes)

	request.Header.Set(field.NewSenderCompID(session.SenderCompID))
	request.Header.Set(field.NewTargetCompID(session.TargetCompID))

	return quickfix.SendToTarget(request.ToMessage(), session)
}
