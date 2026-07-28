package oauth2

import (
	"context"

	"hstserver/model"
	"hstserver/pkg/logger"

	"github.com/goccy/go-json"
)

// invalidateMessage is broadcast to every instance. Either one session or every
// session of a login.
type invalidateMessage struct {
	Sid    string `json:"sid,omitempty"`
	Login  int64  `json:"login,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// Subscribe applies revocations published by the other instances. Losing a
// message costs at most one cache TTL of staleness; correctness never depends
// on delivery.
func (o *OAuth2) Subscribe() {
	ctx := context.Background()
	sub := o.Redis.Client.Subscribe(ctx, ChannelInvalidate)
	defer func() { _ = sub.Close() }()

	ch := sub.Channel()

	for {
		select {
		case <-o.stop:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			var m invalidateMessage
			if err := json.Unmarshal([]byte(msg.Payload), &m); err != nil {
				o.Log.Journal(logger.TypeUser, logger.CodeWarn, "bad invalidate message",
					"error", err.Error())
				continue
			}

			switch {
			case m.Sid != "":
				o.Cache.Invalidate(m.Sid)
			case m.Login != 0:
				o.Cache.InvalidateLogin(m.Login)
			}
		}
	}
}

// publish tells the other instances to drop what we just revoked.
func (o *OAuth2) publish(ctx context.Context, m invalidateMessage) {
	raw, err := json.Marshal(m)
	if err != nil {
		return
	}

	if err := o.Redis.Client.Publish(ctx, ChannelInvalidate, raw).Err(); err != nil {
		o.Log.Journal(logger.TypeUser, logger.CodeWarn, "failed to publish invalidation",
			"error", err.Error())
	}
}

// PackManagerRights turns the 77 columns into the two word bitset the snapshot
// carries, in the column order the migration defines.
func PackManagerRights(m *model.Manager) model.ManagerRights {
	var r model.ManagerRights

	for bit, granted := range map[uint]int32{
		model.MgrRightAdmin:                  m.RightAdmin,
		model.MgrRightManager:                m.RightManager,
		model.MgrRightCfgTime:                m.RightCfgTime,
		model.MgrRightCfgHolidays:            m.RightCfgHolidays,
		model.MgrRightCfgGroups:              m.RightCfgGroups,
		model.MgrRightCfgManagers:            m.RightCfgManagers,
		model.MgrRightCfgRequests:            m.RightCfgRequests,
		model.MgrRightCfgGateways:            m.RightCfgGateways,
		model.MgrRightCfgDatafeeds:           m.RightCfgDatafeeds,
		model.MgrRightCfgReports:             m.RightCfgReports,
		model.MgrRightCfgSymbols:             m.RightCfgSymbols,
		model.MgrRightCfgWebServices:         m.RightCfgWebServices,
		model.MgrRightCfgMessengers:          m.RightCfgMessengers,
		model.MgrRightCfgKyc:                 m.RightCfgKyc,
		model.MgrRightCfgAutomations:         m.RightCfgAutomations,
		model.MgrRightCfgAllocations:         m.RightCfgAllocations,
		model.MgrRightCfgCorporate:           m.RightCfgCorporate,
		model.MgrRightCfgPayments:            m.RightCfgPayments,
		model.MgrRightCfgMails:               m.RightCfgMails,
		model.MgrRightCfgStreaming:           m.RightCfgStreaming,
		model.MgrRightSrvJournals:            m.RightSrvJournals,
		model.MgrRightSrvReports:             m.RightSrvReports,
		model.MgrRightCharts:                 m.RightCharts,
		model.MgrRightEmail:                  m.RightEmail,
		model.MgrRightNews:                   m.RightNews,
		model.MgrRightExport:                 m.RightExport,
		model.MgrRightTechsupport:            m.RightTechsupport,
		model.MgrRightMarket:                 m.RightMarket,
		model.MgrRightAccountant:             m.RightAccountant,
		model.MgrRightAccRead:                m.RightAccRead,
		model.MgrRightAccDetailsName:         m.RightAccDetailsName,
		model.MgrRightAccDetailsLocation:     m.RightAccDetailsLocation,
		model.MgrRightAccDetailsAddress:      m.RightAccDetailsAddress,
		model.MgrRightAccDetailsId:           m.RightAccDetailsId,
		model.MgrRightAccDetailsEmail:        m.RightAccDetailsEmail,
		model.MgrRightAccDetailsPhone:        m.RightAccDetailsPhone,
		model.MgrRightAccDetailsGeneral:      m.RightAccDetailsGeneral,
		model.MgrRightAccTechnical:           m.RightAccTechnical,
		model.MgrRightAccTechModify:          m.RightAccTechModify,
		model.MgrRightAccManager:             m.RightAccManager,
		model.MgrRightAccDelete:              m.RightAccDelete,
		model.MgrRightAccOnline:              m.RightAccOnline,
		model.MgrRightConfirmActions:         m.RightConfirmActions,
		model.MgrRightNotifications:          m.RightNotifications,
		model.MgrRightTradesRead:             m.RightTradesRead,
		model.MgrRightTradesManager:          m.RightTradesManager,
		model.MgrRightTradesDelete:           m.RightTradesDelete,
		model.MgrRightTradesDealer:           m.RightTradesDealer,
		model.MgrRightTradesSupervisor:       m.RightTradesSupervisor,
		model.MgrRightQuotesRaw:              m.RightQuotesRaw,
		model.MgrRightQuotes:                 m.RightQuotes,
		model.MgrRightSymbolDetails:          m.RightSymbolDetails,
		model.MgrRightRiskManager:            m.RightRiskManager,
		model.MgrRightGroupMargin:            m.RightGroupMargin,
		model.MgrRightGroupCommission:        m.RightGroupCommission,
		model.MgrRightReports:                m.RightReports,
		model.MgrRightClientsAccess:          m.RightClientsAccess,
		model.MgrRightClientsCreate:          m.RightClientsCreate,
		model.MgrRightClientsEdit:            m.RightClientsEdit,
		model.MgrRightClientsDelete:          m.RightClientsDelete,
		model.MgrRightClientsKyc:             m.RightClientsKyc,
		model.MgrRightClientsDetailsName:     m.RightClientsDetailsName,
		model.MgrRightClientsDetailsLocation: m.RightClientsDetailsLocation,
		model.MgrRightClientsDetailsAddress:  m.RightClientsDetailsAddress,
		model.MgrRightClientsDetailsId:       m.RightClientsDetailsId,
		model.MgrRightClientsDetailsEmail:    m.RightClientsDetailsEmail,
		model.MgrRightClientsDetailsPhone:    m.RightClientsDetailsPhone,
		model.MgrRightClientsDetailsGeneral:  m.RightClientsDetailsGeneral,
		model.MgrRightDocumentsAccess:        m.RightDocumentsAccess,
		model.MgrRightDocumentsCreate:        m.RightDocumentsCreate,
		model.MgrRightDocumentsEdit:          m.RightDocumentsEdit,
		model.MgrRightDocumentsDelete:        m.RightDocumentsDelete,
		model.MgrRightDocumentsFilesAdd:      m.RightDocumentsFilesAdd,
		model.MgrRightDocumentsFilesDelete:   m.RightDocumentsFilesDelete,
		model.MgrRightCommentsAccess:         m.RightCommentsAccess,
		model.MgrRightCommentsCreate:         m.RightCommentsCreate,
		model.MgrRightCommentsDelete:         m.RightCommentsDelete,
	} {
		if granted == 1 {
			r = r.Set(bit)
		}
	}

	return r
}
