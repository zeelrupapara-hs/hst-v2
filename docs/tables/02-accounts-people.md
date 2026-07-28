# MT5 SQL Export — Accounts & People Domain

Source: MetaTrader 5 Help, *Platform Components > Backup Server > SQL Export*
(`sql_mt5_accounts.htm`, `sql_mt5_users.htm`, `sql_mt5_clients.htm`, `sql_mt5_documents.htm`,
`sql_mt5_managers.htm`, `sql_mt5_email.htm`, `sql_mt5_allocations.htm`,
`sql_mt5_allocations_agreements.htm`).

## MT5 conceptual split (important)

MT5 deliberately separates the *person*, the *login*, the *live trading state*, and the
*back-office operator*. These are four different tables, not one:

- **`mt5_clients`** — the **KYC person / legal entity**. One record per real client
  (individual, corporate, fund). Holds identity, compliance approval, personal/company
  details, contact, address, experience, lead attribution. PK `ClientID`. A client may
  own many trading logins.
- **`mt5_users`** — the **login / trading account record** (the "account database").
  One record per trading login: group membership, permissions, leverage, credentials
  metadata, denormalized balance/equity snapshots, agent, and back-reference `ClientID`.
  PK `Login`.
- **`mt5_accounts`** — the **live trade state** of a login: balance, credit, margin,
  equity, floating P/L, blocked amounts, assets/liabilities. Strictly **1:1 with
  `mt5_users`** — same PK `Login`. `mt5_users` carries slowly-changing/reference data,
  `mt5_accounts` carries fast-moving money state.
- **`mt5_managers`** — the **back-office login** (administrator/manager terminal
  operator), with a large flat matrix of permission flags. PK `Login`, taken from the
  user login the manager account was created on. Managers are referenced *by login* from
  client/document approval fields.

Supporting configuration tables in this domain: `mt5_documents` (KYC document metadata
per client), `mt5_email` (mail server configurations), `mt5_allocations` +
`mt5_allocations_agreements` (self-service account-opening groups and their agreements).

---

## mt5_accounts

Live state of trading accounts — the money/margin snapshot for each login. 1:1 with `mt5_users`.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Login | Integer | **PK**; FK → `mt5_users.Login` | | The login of the client, to whom the trading account belongs. |
| CurrencyDigits | Integer | | | The number of digits after the decimal point in the account deposit currency. |
| Balance | Float | | | The balance of a trade account. |
| Credit | Float | | | The current amount of credit given to an account. |
| Margin | Float | | | The current margin of the account. |
| MarginFree | Float | | | The free margin of an account. |
| MarginLevel | Float | | | The margin level as a percentage. Calculated as a percentage of the current account equity (Equity) to the margin volume (Margin) — i.e. `MarginLevel = Equity / Margin * 100%`. |
| MarginLeverage | Integer | | | Margin leverage. |
| MarginInitial | Float | | | The current size of the initial margin of positions on a trading account. |
| MarginMaintenance | Float | | | The current size of the maintenance margin of positions on a trading account. |
| Profit | Float | | | The size of the current profit for all open positions. |
| Storage | Float | | | The current size of swaps charged for open positions on the account. |
| Floating | Float | | | The size of floating profit/loss of open positions on the account. Calculated as the sum of **Profit, Storage and Commission of open positions** on the account. |
| Equity | Float | | | The account equity calculated as a sum of **Balance, Credit and Floating**. |
| BlockedCommission | Float | | | The amount of the **standard commission locked** on the account, which has been accumulated during the day/month (not yet charged). |
| BlockedProfit | Float | | | The amount of **intraday profit locked** on the account. |
| Assets | Float | | | The current total amount of assets on a trading account. |
| Liabilities | Float | | | The current total amount of liabilities on a trading account. |

---

## mt5_users

The account database: one record per trading login (credentials, group, permissions, personal copy, financial snapshots).

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Login | Integer | **PK** | | The login of a user. |
| Timestamp | Integer | | | Unique record within the table, used for internal purposes of MetaTrader 5 servers. A change of timestamp means the record has been changed. |
| TimestampTrade | Integer | | | Unique record within the table, used for internal purposes of MetaTrader 5 servers. A change of timestamp means the record has been changed. |
| Group | String | FK → `mt5_groups` (group name) | | User group. |
| CertSerialNumber | Integer | | | The number of a last used certificate for user authorization. |
| Rights | Integer | | **EnUsersRights** (doc: `EnUserRights`) | Flags of the user's permissions. Passed using a value of the EnUsersRights enumeration (sum of values of appropriate flags). |
| Registration | DateTime | | | Time of a client record generation, `YYYY-MM-DD HH:MM:SS`. |
| LastAccess | DateTime | | | Date of the last connection using an account, `YYYY-MM-DD HH:MM:SS`. Not updated in real time (traffic/load saving): only updated on connect if more than 24 hours passed since the previous connection. |
| LastPassChange | DateTime | | | The date of the last password change. |
| LastIP | String | | | The IP address from which the user last connected to the server. |
| Name | String | | | The name of the user. **Obsolete field.** |
| FirstName | String | | | The first name of the client. |
| LastName | String | | | The last name of the client. |
| MiddleName | String | | | The middle name of the client. |
| Company | String | | | The name of user's company. |
| Account | String | | | The number of a user's account in an external bank. |
| Country | String | | | The user's country of residence. |
| Language | Integer | | | User's language in the LANGID format used in MS Windows (value from Prim.lang.identifier). |
| ClientID | Integer | FK → `mt5_clients.ClientID` | | The identifier of the client to whom the trading account corresponds. |
| City | String | | | The user's city of residence. |
| State | String | | | The user's state (region) of residence. |
| ZIPCode | String | | | The user's zip code. |
| Address | String | | | The address of the user. |
| Phone | String | | | The user's phone number. |
| EMail | String | | | The email address of the user. |
| ID | String | | | The number of a user's identity document. |
| Status | String | | | Client's status. |
| Comment | String | | | A comment to the user. |
| Color | COLORREF | | | The color of the user — the color of the user's requests shown when handling requests via the manager terminal. |
| PhonePassword | String | | | The user's phone password. |
| Leverage | Integer | | | The size of a user's leverage. |
| Agent | Integer | FK → `mt5_users.Login` (self-ref) | | Agent account number of the user. |
| Balance | Float | | | The current balance of a user. |
| Credit | Float | | | The current amount of funds credited to the user. |
| InterestRate | Float | | | The amount accrued for the current month calculated based on the annual interest rate. |
| CommissionDaily | Float | | | The amount of commissions charged from the user for a day. |
| CommissionMonthly | Float | | | The total amount of commissions charged from the user for the current month. |
| BalancePrevDay | Float | | | The value of the user's balance as of the end of the previous day. |
| BalancePrevMonth | Float | | | The value of a user's balance as of the end of the previous trading month. |
| EquityPrevDay | Float | | | The user's equity as of the end of the previous day. |
| EquityPrevMonth | Float | | | The value of the user's equity as of the end of the previous trading month. |
| TradeAccounts | String | | | Account numbers in external trading systems and gateway identifiers used for working with those systems. Format: `gateway_ID=account_number\|gateway_ID=account_number...` |
| MQID | String | | | MetaQuotes ID of the user. |
| LeadCampaign | String | | | Name of the marketing campaign a client was attracted by. |
| LeadSource | String | | | Lead source (address of the website a client has come from). |
| ApiData | String | | | User data which can be added via MetaTrader 5 API. Sample: `[{pos:0,app_id:1,valInt:500,valUInt:500,valDbl:0.00000000}]` — specifies user data index, ID of the application that added it, and data of three types: Int, UInt and double. Up to 16 such entries. |
| LimitOrders | Integer | | | The maximum number of active (placed) pending orders allowed on the account. |
| LimitPositions | Integer | | | Maximum value of open positions allowed on the account. |

> Note: login-level flags (e.g. password change requirements, certificate/OTP requirements)
> are conveyed via the **EnUsersLoginFlags** enumeration in the Manager API; the SQL export
> exposes the permission mask through `Rights` (**EnUsersRights**).

---

## mt5_clients

The KYC person or legal entity record (CRM side). One record per real client; may own many logins.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| ClientID | Integer | **PK** | | Unique entry ID. |
| Timestamp | Integer | | | Unique value within the table, used by MT5 servers internally. A changed Timestamp means the record changed. |
| ClientType | Integer | | **EnClientType** (inline) | Client type: 0 — Not specified; 1 — Individual; 2 — Corporate; 3 — Fund. |
| ClientStatus | Integer | | **EnClientStatus** (inline) | Client status: 0 — Not registered; 1 — Registered; 2 — Not interested; 3 — Not completed; 4 — Completed; 5 — Information; 6 — Rejected; 7 — Approved; 8 — Financed; 9 — Active; 10 — Inactive; 11 — Suspended; 12 — Closed; 13 — Deleted. |
| AssignedManager | Integer | FK → `mt5_managers.Login` | | The login of the assigned manager. |
| DateCreated | Integer | | | Client creation date. |
| DateModified | Integer | | | Client's last modification date. |
| Comment | String | | | A comment to the client. |
| ComplianceApprovedBy | Integer | FK → `mt5_managers.Login` | | The login of the manager by whom the client was approved. |
| ComplianceClientCategory | String | | | Client compliance category. Currently not used. |
| ComplianceDateApproval | DateTime | | | The date when the client was approved, `YYYY-MM-DD HH:MM:SS`. |
| ComplianceDateTermination | DateTime | | | The date when the provision of services to the client was discontinued, `YYYY-MM-DD HH:MM:SS`. |
| LeadCampaign | String | | | *(Per docs)* The website from which the client came (lead source). |
| LeadSource | String | | | *(Per docs)* The name of the marketing campaign, as a result of which the client came (lead campaign). |
| Introducer | String | FK → `mt5_users.Login` | | The login of the user by whom the client was introduced. |
| PersonTitle | String | | | Client's title. |
| PersonName | String | | | First name and last name. |
| PersonMiddleName | String | | | Middle name. |
| PersonBirthDate | DateTime | | | Date of birth, `YYYY-MM-DD HH:MM:SS`. |
| PersonCitizenship | String | | | Citizenship. |
| PersonGender | Integer | | **EnClientGender** (inline) | Gender: 0 — Not specified; 1 — Male; 2 — Female. |
| PersonTaxID | String | | | Client's Tax ID. |
| PersonDocumentType | String | | | Type of document: passport, driver's license, etc. |
| PersonDocumentNumber | String | | | Document number. |
| PersonDocumentDate | DateTime | | | Document issue date, `YYYY-MM-DD HH:MM:SS`. |
| PersonDocumentExtra | String | | | Additional document information. |
| PersonEmployment | Integer | | **EnClientEmployment** (inline) | Employment status: 0 — Unemployed; 1 — Employed; 2 — Entrepreneur or self-employed; 3 — Retired; 4 — Student; 5 — Other. |
| PersonIndustry | Integer | | **EnClientIndustry** (inline) | Employment area: 0 — Not specified; 1 — Agriculture, Food and Natural Resources; 2 — Architecture and Construction; 3 — Business Administration and Management; 4 — Art, Audio/Video Technology and Communication; 5 — Education and Training; 6 — State and Administrative Management; 7 — Health; 8 — Tourism and Hospitality; 9 — Information Technology; 10 — Legal and Public Safety, Correction and Protection Services; 11 — Manufacturing; 12 — Marketing and Sales; 13 — Science and Technology; 14 — Engineering and Mathematics; 15 — Transportation, Distribution and Logistics; 16 — Other. |
| PersonEducation | Integer | | **EnClientEducation** (inline) | Education: 0 — Not specified; 1 — Secondary; 2 — Bachelor's degree or equivalent; 3 — Master's degree or equivalent; 4 — PhD or equivalent; 5 — Other. |
| PersonWealthSource | Integer | | **EnClientWealthSource** (inline) | Source of income: 0 — Employment/business activity; 1 — Savings or investments; 2 — Gift or inheritance; 3 — Other. |
| PersonAnnualIncome | Float | | | Annual income. |
| PersonNetWorth | Float | | | Net assets. |
| PersonAnnualDeposit | Float | | | Annual deposit. |
| CompanyName | String | | | Company name. |
| CompanyRegNumber | String | | | Company registration number. |
| CompanyRegDate | String | | | Company registration date. |
| CompanyRegAuthority | String | | | Company registration authority. |
| CompanyVat | String | | | VAT number. |
| CompanyLei | String | | | LEI number for EMIR reports. |
| CompanyLicenseNumber | String | | | Company license number. |
| CompanyLicenseAuthority | String | | | Licensing authority. |
| CompanyCountry | String | | | Country of incorporation. |
| CompanyAddress | String | | | Company's legal address. |
| CompanyWebsite | String | | | Company's website. |
| ContactPreferred | Integer | | **EnClientContactPreferred** (inline) | Preferred method of communication: 0 — Not specified; 1 — Email; 2 — Telephone; 3 — SMS; 4 — Messenger. |
| ContactLanguage | String | | | Client's language. |
| ContactEmail | String | | | Client's email. |
| ContactPhone | String | | | Phone number. |
| ContactMessengers | String | | | Messengers. |
| ContactSocialNetworks | String | | | Accounts in social networks. |
| ContactLastDate | DateTime | | | Last contact date, `YYYY-MM-DD HH:MM:SS`. |
| AddressCountry | String | | | Client's country. |
| AddressPostcode | String | | | Client's postal code. |
| AddressStreet | String | | | Client's address. |
| AddressState | String | | | State/region of residence. |
| AddressCity | String | | | City. |
| ExperienceFX | Integer | | | Forex trading experience, number of years. |
| ExperienceCFD | Integer | | | CFD trading experience, number of years. |
| ExperienceFutures | Integer | | | Futures trading experience, number of years. |
| ExperienceStocks | Integer | | | Stock trading experience, number of years. |
| ClientOrigin | Integer | | **EnClientOrigin** (inline) | How the client record was created: 0 — manually; 1 — based on a demo account; 2 — based on a contest account; 3 — based on a preliminary account; 4 — based on a real account. |
| ClientOriginLogin | Integer | FK → `mt5_users.Login` | | The number of the account based on which the client record was created. |

> Client-level permission flags are represented by the **EnClientRights** enumeration in
> the Manager API; the SQL export surfaces status/type via `ClientStatus`/`ClientType`.

---

## mt5_documents

Metadata for client KYC documents. **Only document data is exported to SQL — the document files themselves are not exported.**

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| DocumentID | Integer | **PK** | | Unique document ID. |
| Timestamp | Integer | | | Unique value within the table, used by MT5 servers internally. A changed Timestamp means the record changed. |
| RelatedClient | Integer | FK → `mt5_clients.ClientID` | | The ID of the client to whom the document belongs. Corresponds to ClientID from the `mt5_clients` table. |
| ApprovedDate | DateTime | | | Document approval date, `YYYY-MM-DD HH:MM:SS`. |
| ApprovedBy | Integer | FK → `mt5_managers.Login` | | The login of the manager by whom the document was approved. |
| DateIssue | DateTime | | | Document issue date, `YYYY-MM-DD HH:MM:SS`. |
| DateExpiration | DateTime | | | Document expiration date, `YYYY-MM-DD HH:MM:SS`. |
| DocumentType | Integer | | **EnDocumentType** (inline) | Document type: 0 — Other; 1 — Proof of identity; 2 — Proof of address; 3 — Registration address; 4 — CEO's ID document; 5 — Certificate of Registration; 6 — Certificate of Directors; 7 — Certificate of good standing. |
| DocumentName | String | | | Document name. |
| DocumentComment | String | | | Comment to the document. |
| DocumentStatus | Integer | | **EnDocumentStatus** (inline) | Document status: 0 — New; 1 — Approved; 2 — Rejected; 3 — Archived; 4 — Deleted. |

---

## mt5_managers

Back-office (administrator/manager terminal) logins and their full permission matrix.
All `Right_*` columns are Integer flags: **1 = permission granted, 0 = no permission**
(the flat expansion of the manager access-rights / **EnManagerRights** flag set).

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Login | Integer | **PK**; FK → `mt5_users.Login` | | The user login, based on which the manager account is created. |
| Timestamp | Integer | | | Unique value within the table, used by MT5 servers internally. A changed Timestamp means the record changed. |
| Name | String | | | The name of the manager. |
| Mailbox | String | | | The name of the manager's mailbox in the internal mailing system. |
| Server | Integer | FK → trade server ID | | The ID of the trade server to which the manager belongs. |
| RequestLimitLogs | Integer | | **EnManagerRequestLimit** (inline) | Time period of system logs available to a manager: 0 — unlimited; 1 — 1 month; 2 — 3 months; 3 — 6 months; 4 — 1 year; 5 — 2 years; 6 — 3 years. |
| RequestLimitReports | Integer | | **EnManagerRequestLimit** (inline) | Time period of reports available to a manager: 0 — unlimited; 1 — 1 month; 2 — 3 months; 3 — 6 months; 4 — 1 year; 5 — 2 years; 6 — 3 years. |
| Groups | String | FK → `mt5_groups` (comma-separated masks) | | The list of groups processed by the manager, comma-separated, e.g. `"demo\forex-netting,demo\forex-netting"`. |
| Access | String | | | The list of IP addresses from which a manager is allowed to connect, e.g. `192.168.0.1-192.168.0.10,192.168.0.12-192.168.0.20`. |
| Right_Admin | Integer | | manager access flag | Connection using the administrator terminal. |
| Right_Manager | Integer | | manager access flag | Connection using the manager terminal. |
| Right_Cfg_Servers | Integer | | manager access flag | Network configuration. |
| Right_Cfg_Access | Integer | | manager access flag | Configuration of the list of IP access. |
| Right_Cfg_Time | Integer | | manager access flag | Configuration of the server working time. |
| Right_Cfg_Holidays | Integer | | manager access flag | Configuration of holidays. |
| Right_Cfg_Groups | Integer | | manager access flag | Configuration of groups. |
| Right_Cfg_Managers | Integer | | manager access flag | Configuration of manager rights. |
| Right_Cfg_Requests | Integer | | manager access flag | Configuration of the routing table. |
| Right_Cfg_Gateways | Integer | | manager access flag | Configuration of gateways. |
| Right_Cfg_Plugins | Integer | | manager access flag | Configuration of plugins. |
| Right_Cfg_Datafeeds | Integer | | manager access flag | Configuration of data feeds. |
| Right_Cfg_Reports | Integer | | manager access flag | Configuration of reports. |
| Right_Cfg_Symbols | Integer | | manager access flag | Configuring the symbols. |
| Right_Cfg_Hst_Sync | Integer | | manager access flag | Configuration of synchronization. |
| Right_Cfg_ECN | Integer | | manager access flag | ECN configuration. |
| Right_Cfg_VPS | Integer | | manager access flag | Configuring Sponsored VPS for traders. |
| Right_Cfg_Web_Services | Integer | | manager access flag | Configuring integration with web services: SSL certificates and addresses for callback requests. |
| Right_Cfg_Funds | Integer | | manager access flag | Configuring investment funds in the Administrator terminal. |
| Right_Cfg_Messengers | Integer | | manager access flag | Configuring integration with SMS providers and messengers in the Administrator terminal. |
| Right_Cfg_KYC | Integer | | manager access flag | Configuring integration with KYC services in the Administrator terminal. |
| Right_Cfg_Automations | Integer | | manager access flag | Configuring automatic actions for specified scenarios in the Administrator terminal. |
| Right_Cfg_Allocations | Integer | | manager access flag | Accessing the Allocations section of the Administrator terminal (groups in which traders can open demo and preliminary real accounts directly from client terminals). |
| Right_Cfg_Corporate | Integer | | manager access flag | Access to corporate link settings. |
| Right_Cfg_Payments | Integer | | manager access flag | Configuring integration with payment systems. |
| Right_Cfg_Mails | Integer | | manager access flag | Configuring integration with email services in the Administrator terminal. |
| Right_Cfg_Streaming | Integer | | manager access flag | Configuring data streaming in external systems, such as Apache Kafka. |
| Right_Srv_Journals | Integer | | manager access flag | Access to server journals. |
| Right_Srv_Reports | Integer | | manager access flag | Receiving automatic server reports. |
| Right_Charts | Integer | | manager access flag | Editing history data on the server. |
| Right_Email | Integer | | manager access flag | Sending internal emails. |
| Right_News | Integer | | manager access flag | Permission to send newsletters. An administrator or manager can only send newsletters if his or her account belongs to a group created on the main trade server. |
| Right_Export | Integer | | manager access flag | Ability to copy to the clipboard and export account, order and other data from Administrator/Manager terminals to external files. |
| Right_Techsupport | Integer | | manager access flag | Access to the technical support tab in the administrator and manager terminals. **Obsolete, no longer used.** |
| Right_Market | Integer | | manager access flag | Permission to access the Market of applications in the MT5 Administrator. **Obsolete, no longer used.** |
| Right_Accountant | Integer | | manager access flag | Permission to work with funds on accounts. |
| Right_Acc_Read | Integer | | manager access flag | Access to accounts. |
| Right_Acc_Details_Name | Integer | | manager access flag | Access to name details in accounts. |
| Right_Acc_Details_Location | Integer | | manager access flag | Access to location data in accounts: country, city, region, zip code. |
| Right_Acc_Details_Address | Integer | | manager access flag | Access to address details in accounts. |
| Right_Acc_Details_ID | Integer | | manager access flag | Access to data on document numbers in accounts. |
| Right_Acc_Details_EMail | Integer | | manager access flag | Access to email details in accounts. |
| Right_Acc_Details_Phone | Integer | | manager access flag | Access to phone details in accounts. |
| Right_Acc_Details_General | Integer | | manager access flag | Access to other data in accounts (language, status, comment, MetaQuotes ID, etc.). |
| Right_Acc_Technical | Integer | | manager access flag | Combined with the "Show to regular managers" permission, eases work with testing/technical accounts: disable visibility for regular managers on technical accounts, then disable access to technical accounts for managers who do not configure the platform. |
| Right_Acc_Tech_Modify | Integer | | manager access flag | Allows enabling/disabling the "Show to regular managers" and "Include in server reports" options for a trading account. Without it the manager sees these options read-only. |
| Right_Acc_Manager | Integer | | manager access flag | Account editing. |
| Right_Acc_Delete | Integer | | manager access flag | Deleting client accounts via administrator/manager terminals and Manager API. Requires `Right_Acc_Manager`. |
| Right_Acc_Online | Integer | | manager access flag | Getting the current client connections. |
| Right_Confirm_Actions | Integer | | manager access flag | By default the manager terminal shows a confirmation dialog for balance operations and closing multiple orders (manager must type a random character sequence). If disabled, those actions execute immediately without confirmation. |
| Right_Notifications | Integer | | manager access flag | Permission to send push notifications to clients' mobile devices from the manager terminal, based on MetaQuotes ID (unique user identifier obtained by installing MT5 Mobile for iPhone/Android). |
| Right_Trades_Read | Integer | | manager access flag | Viewing trading orders, deals and positions. Gates `Right_Trades_Manager` and `Right_Trades_Dealer`. |
| Right_Trades_Manager | Integer | | manager access flag | Changing any fields of orders, deals and positions in the administrator terminal, and changing position open prices in the manager terminal. |
| Right_Trades_Delete | Integer | | manager access flag | Deleting any orders, deals and positions via administrator/manager terminals and Manager API. Requires `Right_Trades_Manager`. |
| Right_Trades_Dealer | Integer | | manager access flag | The possibility to perform trading and dealing operations in the manager terminal. |
| Right_Trades_Supervisor | Integer | | manager access flag | Allows viewing the entire queue of requests from client groups available to the manager and tracking request processing by other dealers ("Supervisor" mode, without connecting as a dealer). After connecting as a dealer, only requests routed to him/her are visible. |
| Right_Quotes_Raw | Integer | | manager access flag | Enables the "Show raw quotes" command in the Market Watch context menu — view quotes without the spread difference settings of the manager group. |
| Right_Quotes | Integer | | manager access flag | Permission to throw in quotes. |
| Right_Symbol_Details | Integer | | manager access flag | Permission to change spread and execution mode. |
| Right_Risk_Manager | Integer | | manager access flag | Permission to receive information about client's aggregate positions and company's coverage positions. |
| Right_Group_Margin | Integer | | manager access flag | Permission to configure margin for groups in MT5 Manager. |
| Right_Group_Commission | Integer | | manager access flag | Permission to configure commissions for groups in MT5 Manager. |
| Right_Reports | Integer | | manager access flag | Permission to request and receive various reports on client operations. |
| Right_Finteza_Access | Integer | | manager access flag | Access to the Finteza Analytics section. |
| Right_Finteza_Websites | Integer | | manager access flag | View Finteza data relating to websites in the Analytics section of the Manager terminal. |
| Right_Finteza_Campaigns | Integer | | manager access flag | View Finteza data relating to marketing campaigns in the Analytics section of the Manager terminal. |
| Right_Finteza_Reports | Integer | | manager access flag | Currently not used. |
| Right_Clients_Access | Integer | | manager access flag | Access to the Clients section in the Administrator and Manager terminals. |
| Right_Clients_Create | Integer | | manager access flag | Permission to create new client records manually. |
| Right_Clients_Edit | Integer | | manager access flag | Permission to edit client data, except for documents. |
| Right_Clients_Delete | Integer | | manager access flag | Permission to delete client records. |
| Right_Clients_KYC | Integer | | manager access flag | Permission to launch automated validation of client data via integrated KYC services (from Manager/Administrator terminals and API). |
| Right_Clients_Details_Name | Integer | | manager access flag | Access to name details in clients. |
| Right_Clients_Details_Location | Integer | | manager access flag | Access to location data in clients: country, city, region, zip code. |
| Right_Clients_Details_Address | Integer | | manager access flag | Access to address details in clients. |
| Right_Clients_Details_Id | Integer | | manager access flag | Access to data on document numbers in clients. |
| Right_Clients_Details_Email | Integer | | manager access flag | Access to email details in clients. |
| Right_Clients_Details_Phone | Integer | | manager access flag | Access to phone details in clients. |
| Right_Clients_Details_General | Integer | | manager access flag | Access to other data in clients (language, status, Lead Source, Lead Campaign, etc.). |
| Right_Documents_Access | Integer | | manager access flag | Permission to view client documents. |
| Right_Documents_Create | Integer | | manager access flag | Permission to add general information about documents in client records. |
| Right_Documents_Edit | Integer | | manager access flag | Permission to edit general information about documents in client records. |
| Right_Documents_Delete | Integer | | manager access flag | Permission to delete general information about documents from client records. |
| Right_Documents_Files_Add | Integer | | manager access flag | Permission to add document files in client records. |
| Right_Documents_Files_Delete | Integer | | manager access flag | Permission to delete document files from client records. |
| Right_Comments_Access | Integer | | manager access flag | Permission to read comments to clients and their documents. |
| Right_Comments_Create | Integer | | manager access flag | Permission to write comments to clients and their documents. |
| Right_Comments_Delete | Integer | | manager access flag | Permission to delete comments to clients and their documents. |
| Right_Admin_Computer | Integer | | manager access flag | Access to the server machine administration menu in the Network section of the Administrator terminal. |
| Right_Subscriptions_View | Integer | | manager access flag | Permission to view existing settings in the Subscriptions section of the Administrator terminal, plus access to subscription statistics. |
| Right_Subscriptions_Edit | Integer | | manager access flag | Permission to create, edit and remove settings in the Subscriptions section of the Administrator terminal. |
| Right_Payments_Access | Integer | | manager access flag | Permission to view current and processed payments and payment accounts. |
| Right_Payments_Process | Integer | | manager access flag | Permission to confirm and reject payments processed manually via the Manager terminal. |
| Right_Payments_Edit | Integer | | manager access flag | Permission to edit current and processed payments and payment accounts. |
| Right_Payments_Delete | Integer | | manager access flag | Permission to delete current and processed payments and payment accounts. |
| Right_Ultency_Access | Integer | | manager access flag | Permission to view Ultency settings, including provider gallery, provider symbols, aggregated symbols, translations and routing. |
| Right_Ultency_Edit | Integer | | manager access flag | Permission to edit Ultency settings (as listed above), including adding new liquidity providers. |
| Right_Ultency_Delete | Integer | | manager access flag | Permission to delete Ultency settings (as listed above). |
| Right_Ultency_Servicedesk | Integer | | manager access flag | Permission to use the built-in communication system. |

---

## mt5_email

Mail server configurations (used for outgoing platform email, e.g. account-opening confirmations).

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Name | String | **PK** (config name); referenced by `mt5_allocations.ConfirmationEmail` | | Mail server configuration name. |
| Timestamp | Integer | | | Unique value within the table, used by MT5 servers internally. A changed Timestamp means the record was modified. |
| SenderMail | String | | | The email address from which emails are sent via the mail server configuration. |
| SenderName | String | | | The sender name in the mail server configuration. |
| Server | String | | | The SMTP server address in the mail server configuration. |
| Login | String | | | The SMTP login in the mail server configuration. |
| Flags | Integer | | **EnMailFlags** (doc: `EnFlags`) | Advanced mail server settings, transferred via the mail `EnFlags` enumeration. |

---

## mt5_allocations

Account allocation settings — the groups in which traders can self-open demo/preliminary-real accounts from the client terminal.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Allocation_ID | Integer | **PK** | | A unique configuration identifier for efficient retrieval from the database. Assigned automatically upon export. |
| Timestamp | Integer | | | Unique value within the table, used by MT5 servers internally. A changed Timestamp means the record was modified. |
| Group | String | FK → `mt5_groups` (group name) | | The group in which the accounts requested through terminals will be opened. |
| Description | String | | | Group description displayed in the "Account type" field in client terminals. |
| Company | String | | | The name of the company whose terminals will have access to this group. |
| Currency | String | | | Deposit currency in the specified group. |
| CurrencyDigits | Integer | | | Deposit currency accuracy. |
| AccountType | Integer | | **EnAllocationAccountType** (inline) | Group's risk management type: 0 — Retail Forex, CFD, Futures; 1 — Stock Exchange, based on margin discount rates; 2 — Retail Forex, CFD, Futures with hedging. |
| DefaultLeveleage | Integer | | | Default leverage in the specified group. *(Field name is misspelled in the MT5 schema — keep as-is.)* |
| DefaultDeposit | Integer | | | Default initial deposit in the specified group. |
| Flags | Integer | | **EnAllocationFlags** (doc: `EnFlags`) | Additional account allocation settings, transferred via the allocation `EnFlags` enumeration. |
| Leverages | String | | | The list of available leverage options which can be selected when opening an account in this group. |
| Countries | String | | | A list of countries where account opening will be available for the specified group. |
| ConfirmationEmail | String | FK → `mt5_email.Name` | | The mail server that will be used to verify email addresses when opening accounts in the specified group. |

---

## mt5_allocations_agreements

The list of agreements (terms/documents the trader must accept) used in the account allocation settings.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Agreement_ID | Integer | **PK** | | A unique configuration identifier for efficient retrieval from the database. Assigned automatically upon export. |
| Allocation_ID | Integer | FK → `mt5_allocations.Allocation_ID` | | The identifier of the account allocation configuration to which the agreement belongs. |
| Timestamp | Integer | | | Unique value within the table, used by MT5 servers internally. A changed Timestamp means the record was modified. |
| Type | Integer | | **EnCaptionType** | Agreement type, transferred using the EnCaptionType enumeration. |
| CaptionCustom | String | | | Title of the user agreement. |
| URL | String | | | The link where the agreement can be accessed. |
| Flags | Integer | | **EnAllocationFlags** (doc: `EnFlags`) | Additional account allocation settings, transferred via the `EnFlags` enumeration. |
