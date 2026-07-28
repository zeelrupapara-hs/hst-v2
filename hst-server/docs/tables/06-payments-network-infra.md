# MT5 SQL Export — Domain 06: Payments, Corporate, Network & Security

> **Scope note.** Everything in this domain is *broker infrastructure*, not trading mechanics:
> PSP/payment-wallet configuration and payment transactions, corporate-link (White Label) settings,
> physical server topology (trade / history / backup / access / Anti-DDoS servers), firewall rules,
> and messenger / SMS provider configuration.
> **Priority: LOW for a v1 trading-engine build.** None of these tables are on the order/deal/position
> execution path. Domains 01-04 (accounts, groups, symbols, orders/deals/positions) come first.
> `mt5_payments` / `mt5_payment_history` become relevant only once a deposit/withdrawal flow is built;
> the `mt5_network_*` family is operational telemetry emitted by the MT5 backup server itself and would
> normally have no analogue in a from-scratch engine.

---

## mt5_payments

Active (in-flight) payment transactions: deposits/withdrawals routed through a payment provider wallet.
Very wide table — most columns are optional KYC / bank / crypto / beneficiary detail fields populated only by the PSPs that require them.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Payment_ID | Integer | PK | - | A unique identifier of the parameter. |
| Timestamp | Integer | - | - | Unique value within the table; used by MT5 servers internally. A changed Timestamp means the record was modified. |
| Action | Integer | - | EnPaymentAction | Transaction type. |
| State | Integer | - | EnPaymentState | Transaction status. |
| TimeCreated | DateTime | - | - | Transaction creation time. |
| TimeDone | DateTime | - | - | Transaction execution time. |
| WalletType | String | - | - | Name of the payment provider. |
| WalletName | String | - | - | Name of the wallet via which the transaction is performed. |
| Description | String | - | - | Payment description. |
| IP | String | - | - | IP address from which the transaction is performed. |
| Locale | String | - | - | Name of the region from which the transaction is performed. |
| PolingTime | DateTime | - | - | Time of the last data request from the payment system. |
| PolingCounter | Integer | - | - | Number of data requests from the payment system. |
| Login | Integer | FK -> mt5_users.Login | - | Account number. |
| Manager | Integer | FK -> mt5_users.Login | - | Login of the manager who processed the payment. |
| DealID | Integer | FK -> mt5_deals.Deal | - | Ticket of the deal used to post the payment to the trading account. |
| MailID | Integer | - | - | User's email address. |
| TaskID | Integer | - | - | Identifier of the associated task. Currently unused. |
| Invoice | String | - | - | Account number. |
| UserCurrency | String | - | - | Account deposit currency. |
| UserCurrencyDigits | Integer | - | - | Number of decimal places in the account's deposit currency. |
| UserAmount | Float | - | - | Transaction amount requested by the client. |
| UserCommission | Float | - | - | Commission charged by the broker according to the settings. |
| WalletCurrency | String | - | - | Currency selected by the user in the client terminal when initiating the transaction. |
| WalletCurrencyDigits | Integer | - | - | Number of decimal places in the currency selected by the user. |
| WalletAmount | Float | - | - | Transaction amount on the payment provider's side. |
| WalletCommission | Float | - | - | Payment provider's commission for the transaction. |
| ConversionRate | Float | - | - | Exchange rate used to convert the amount from the client's currency into the payment system's currency (or vice versa). |
| ExternalID | Integer | - | - | Transaction identifier assigned by the payment system. |
| ExernalErrorCode | Integer | - | - | Error code returned by the payment system. |
| ExernalErrorDesc | String | - | - | Error description returned by the payment system. |
| PaymentType | Integer | - | EnPaymentType | Payment method. |
| Flags | Integer | - | EnPaymentFlags | Additional wallet flags. |
| ClientType | Integer | - | EnUsersConnectionTypes | Client type that initiated the payment. |
| CustomerFirstName | String | - | - | Client's first name. |
| CustomerLastName | String | - | - | Client's last name. |
| CustomerBirthDate | DateTime | - | - | Client's date of birth. |
| CustomerEmail | String | - | - | Client's email. |
| CustomerPhone | String | - | - | Client's phone number. |
| CustomerPhoneHome | String | - | - | Client's home phone number. |
| CustomerPhoneWork | String | - | - | Client's work phone number. |
| CustomerFullName | String | - | - | Client's full name. |
| CustomerTIN | String | - | - | Client's TIN. |
| CustomerCPF | String | - | - | Client's CPF (for Portugal). |
| CustomerDocumentType | String | - | - | Client's document type. |
| CustomerDocumentNumber | String | - | - | Client's document number. |
| EwalletAccount | String | - | - | Account in an electronic payment system. |
| EwalletEmail | String | - | - | Mailbox in an electronic payment system. |
| EwalletPhone | String | - | - | Phone number in an electronic payment system. |
| EwalletLogin | String | - | - | Login in an electronic payment system. |
| BankBranch | String | - | - | Bank branch. |
| BankCode | String | - | - | Bank code. |
| BankName | String | - | - | Bank name. |
| BankAccount | String | - | - | Bank account number. |
| BankAccountName | String | - | - | Bank account holder's name. |
| BankAccountType | String | - | - | Bank account type. |
| BankTaxReasonCode | String | - | - | Tax registration reason code. |
| BankAddressLine1 | String | - | - | Account holder's address. |
| BankAddressLine2 | String | - | - | Account holder's address. |
| BankAddressCity | String | - | - | Account holder's city. |
| BankAddressCountry | String | - | - | Account holder's country. |
| BankAddressState | String | - | - | Account holder's region. |
| BankAddressPostcode | String | - | - | Account holder's postal code. |
| BillingAddressLine1 | String | - | - | Billing address. |
| BillingAddressLine2 | String | - | - | Billing address. |
| BillingAddressCity | String | - | - | Billing city. |
| BillingAddressCountry | String | - | - | Billing country. |
| BillingAddressState | String | - | - | Billing region. |
| BillingAddressPostcode | String | - | - | Billing postal code. |
| LivingAddressLine1 | String | - | - | Residential address. |
| LivingAddressLine2 | String | - | - | Residential address. |
| LivingAddressCity | String | - | - | City of residence. |
| LivingAddressCountry | String | - | - | Country of residence. |
| LivingAddressState | String | - | - | Region of residence. |
| LivingAddressPostcode | String | - | - | Postal code at the place of residence. |
| PaymentDetails | String | - | - | Payment description. Used by some systems. |
| PaymentAccountID | Integer | - | - | Account identifier. Used by some systems. |
| PaymentPixKey | String | - | - | Pix payment system key. |
| MobileOperatorName | String | - | - | Mobile operator name. |
| MobileOperatorCode | String | - | - | Mobile operator code. |
| CryptoChain | String | - | - | Crypto chain. |
| CryptoToken | String | - | - | Crypto token. |
| CryptoAddress | String | - | - | Crypto address. |
| CryptoDestination | String | - | - | Crypto destination tag. |
| CryptoTransaction | String | - | - | Crypto transaction. |
| BeneficiaryName | String | - | - | Recipient's name. |
| BeneficiaryAddressLine1 | String | - | - | Recipient's address. |
| BeneficiaryAddressLine2 | String | - | - | Recipient's address. |
| BeneficiaryAddressCity | String | - | - | Recipient's city. |
| BeneficiaryAddressCountry | String | - | - | Recipient's country. |
| BeneficiaryAddressState | String | - | - | Recipient's region. |
| BeneficiaryAddressPostcode | String | - | - | Recipient's postal code. |
| BeneficiaryBankBranch | String | - | - | Recipient bank branch. |
| BeneficiaryBankCode | String | - | - | Recipient bank code. |
| BeneficiaryBankName | String | - | - | Recipient bank name. |
| BeneficiaryBankAccount | String | - | - | Recipient's bank account number. |
| BeneficiaryBankAccountName | String | - | - | Recipient's bank account holder name. |
| BeneficiaryBankAccountType | String | - | - | Recipient's bank account type. |
| BeneficiaryBankTaxReasonCode | String | - | - | Recipient's tax registration reason code. |
| BeneficiaryBankAddressLine1 | String | - | - | Recipient bank address. |
| BeneficiaryBankAddressLine2 | String | - | - | Recipient bank address. |
| BeneficiaryBankAddressCity | String | - | - | Recipient bank city. |
| BeneficiaryBankAddressCountry | String | - | - | Recipient bank country. |
| BeneficiaryBankAddressState | String | - | - | Recipient bank region. |
| BeneficiaryBankAddressPostcode | String | - | - | Recipient bank postal code. |

## mt5_payment_history

Completed/archived payment transactions. Field set is byte-for-byte identical to `mt5_payments`;
records move here once the transaction leaves the active queue.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Payment_ID | Integer | PK | - | A unique identifier of the parameter. |
| Timestamp | Integer | - | - | Unique value within the table; used by MT5 servers internally. A changed Timestamp means the record was modified. |
| Action | Integer | - | EnPaymentAction | Transaction type. |
| State | Integer | - | EnPaymentState | Transaction status. |
| TimeCreated | DateTime | - | - | Transaction creation time. |
| TimeDone | DateTime | - | - | Transaction execution time. |
| WalletType | String | - | - | Name of the payment provider. |
| WalletName | String | - | - | Name of the wallet via which the transaction is performed. |
| Description | String | - | - | Payment description. |
| IP | String | - | - | IP address from which the transaction is performed. |
| Locale | String | - | - | Name of the region from which the transaction is performed. |
| PolingTime | DateTime | - | - | Time of the last data request from the payment system. |
| PolingCounter | Integer | - | - | Number of data requests from the payment system. |
| Login | Integer | FK -> mt5_users.Login | - | Account number. |
| Manager | Integer | FK -> mt5_users.Login | - | Login of the manager who processed the payment. |
| DealID | Integer | FK -> mt5_deals.Deal | - | Ticket of the deal used to post the payment to the trading account. |
| MailID | Integer | - | - | User's email address. |
| TaskID | Integer | - | - | Identifier of the associated task. Currently unused. |
| Invoice | String | - | - | Account number. |
| UserCurrency | String | - | - | Account deposit currency. |
| UserCurrencyDigits | Integer | - | - | Number of decimal places in the account's deposit currency. |
| UserAmount | Float | - | - | Transaction amount requested by the client. |
| UserCommission | Float | - | - | Commission charged by the broker according to the settings. |
| WalletCurrency | String | - | - | Currency selected by the user in the client terminal when initiating the transaction. |
| WalletCurrencyDigits | Integer | - | - | Number of decimal places in the currency selected by the user. |
| WalletAmount | Float | - | - | Transaction amount on the payment provider's side. |
| WalletCommission | Float | - | - | Payment provider's commission for the transaction. |
| ConversionRate | Float | - | - | Exchange rate used to convert the amount from the client's currency into the payment system's currency (or vice versa). |
| ExternalID | Integer | - | - | Transaction identifier assigned by the payment system. |
| ExernalErrorCode | Integer | - | - | Error code returned by the payment system. |
| ExernalErrorDesc | String | - | - | Error description returned by the payment system. |
| PaymentType | Integer | - | EnPaymentType | Payment method. |
| Flags | Integer | - | EnPaymentFlags | Additional wallet flags. |
| ClientType | Integer | - | EnUsersConnectionTypes | Client type that initiated the payment. |
| CustomerFirstName | String | - | - | Client's first name. |
| CustomerLastName | String | - | - | Client's last name. |
| CustomerBirthDate | DateTime | - | - | Client's date of birth. |
| CustomerEmail | String | - | - | Client's email. |
| CustomerPhone | String | - | - | Client's phone number. |
| CustomerPhoneHome | String | - | - | Client's home phone number. |
| CustomerPhoneWork | String | - | - | Client's work phone number. |
| CustomerFullName | String | - | - | Client's full name. |
| CustomerTIN | String | - | - | Client's TIN. |
| CustomerCPF | String | - | - | Client's CPF (for Portugal). |
| CustomerDocumentType | String | - | - | Client's document type. |
| CustomerDocumentNumber | String | - | - | Client's document number. |
| EwalletAccount | String | - | - | Account in an electronic payment system. |
| EwalletEmail | String | - | - | Mailbox in an electronic payment system. |
| EwalletPhone | String | - | - | Phone number in an electronic payment system. |
| EwalletLogin | String | - | - | Login in an electronic payment system. |
| BankBranch | String | - | - | Bank branch. |
| BankCode | String | - | - | Bank code. |
| BankName | String | - | - | Bank name. |
| BankAccount | String | - | - | Bank account number. |
| BankAccountName | String | - | - | Bank account holder's name. |
| BankAccountType | String | - | - | Bank account type. |
| BankTaxReasonCode | String | - | - | Tax registration reason code. |
| BankAddressLine1 | String | - | - | Account holder's address. |
| BankAddressLine2 | String | - | - | Account holder's address. |
| BankAddressCity | String | - | - | Account holder's city. |
| BankAddressCountry | String | - | - | Account holder's country. |
| BankAddressState | String | - | - | Account holder's region. |
| BankAddressPostcode | String | - | - | Account holder's postal code. |
| BillingAddressLine1 | String | - | - | Billing address. |
| BillingAddressLine2 | String | - | - | Billing address. |
| BillingAddressCity | String | - | - | Billing city. |
| BillingAddressCountry | String | - | - | Billing country. |
| BillingAddressState | String | - | - | Billing region. |
| BillingAddressPostcode | String | - | - | Billing postal code. |
| LivingAddressLine1 | String | - | - | Residential address. |
| LivingAddressLine2 | String | - | - | Residential address. |
| LivingAddressCity | String | - | - | City of residence. |
| LivingAddressCountry | String | - | - | Country of residence. |
| LivingAddressState | String | - | - | Region of residence. |
| LivingAddressPostcode | String | - | - | Postal code at the place of residence. |
| PaymentDetails | String | - | - | Payment description. Used by some systems. |
| PaymentAccountID | Integer | - | - | Account identifier. Used by some systems. |
| PaymentPixKey | String | - | - | Pix payment system key. |
| MobileOperatorName | String | - | - | Mobile operator name. |
| MobileOperatorCode | String | - | - | Mobile operator code. |
| CryptoChain | String | - | - | Crypto chain. |
| CryptoToken | String | - | - | Crypto token. |
| CryptoAddress | String | - | - | Crypto address. |
| CryptoDestination | String | - | - | Crypto destination tag. |
| CryptoTransaction | String | - | - | Crypto transaction. |
| BeneficiaryName | String | - | - | Recipient's name. |
| BeneficiaryAddressLine1 | String | - | - | Recipient's address. |
| BeneficiaryAddressLine2 | String | - | - | Recipient's address. |
| BeneficiaryAddressCity | String | - | - | Recipient's city. |
| BeneficiaryAddressCountry | String | - | - | Recipient's country. |
| BeneficiaryAddressState | String | - | - | Recipient's region. |
| BeneficiaryAddressPostcode | String | - | - | Recipient's postal code. |
| BeneficiaryBankBranch | String | - | - | Recipient bank branch. |
| BeneficiaryBankCode | String | - | - | Recipient bank code. |
| BeneficiaryBankName | String | - | - | Recipient bank name. |
| BeneficiaryBankAccount | String | - | - | Recipient's bank account number. |
| BeneficiaryBankAccountName | String | - | - | Recipient's bank account holder name. |
| BeneficiaryBankAccountType | String | - | - | Recipient's bank account type. |
| BeneficiaryBankTaxReasonCode | String | - | - | Recipient's tax registration reason code. |
| BeneficiaryBankAddressLine1 | String | - | - | Recipient bank address. |
| BeneficiaryBankAddressLine2 | String | - | - | Recipient bank address. |
| BeneficiaryBankAddressCity | String | - | - | Recipient bank city. |
| BeneficiaryBankAddressCountry | String | - | - | Recipient bank country. |
| BeneficiaryBankAddressState | String | - | - | Recipient bank region. |
| BeneficiaryBankAddressPostcode | String | - | - | Recipient bank postal code. |

## mt5_pay_wallets

Payment wallet (PSP) configurations: one row per configured payment provider connection on a trade server.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| PaymentWallet_ID | Integer | PK | - | A unique identifier of the wallet configuration. |
| Timestamp | Integer | - | - | Unique value within the table; used internally. A changed Timestamp means the record was modified. |
| Name | String | - | - | Wallet configuration name. |
| Server | Integer | FK -> mt5_network.Login | - | Identifier of the trade server with which the configuration is associated. |
| Provider | String | - | - | Provider name. |
| Enable | Integer | - | inline | Wallet operating mode: 0 — disabled, 1 — enabled. |
| Login | String | - | - | Username or token used to connect to the payment provider. |
| Actions | Integer | - | EnPaymentProviderActionFlags | Transaction types available for the wallet. |
| Flags | Integer | - | EnFlags | Additional wallet parameters. |

## mt5_pay_wallet_params

Free-form typed key/value parameters attached to a wallet configuration (provider-specific settings).

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Param_ID | Integer | PK | - | A unique identifier of the parameter. |
| PaymentWallet_ID | Integer | FK -> mt5_pay_wallets.PaymentWallet_ID | - | Wallet configuration ID to which the parameters apply. |
| Type | Integer | - | inline | Parameter type: 0 — string; 1 — integer; 2 — floating-point number; 3 — time; 4 — date; 5 — date and time; 6 — list of groups; 7 — list of symbols; 8 — Boolean; 9 — color. |
| Name | String | - | - | Parameter name. |
| Value | String | - | - | Parameter value. |

## mt5_pay_wallet_groups

Account-group availability list for a wallet configuration.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Group_ID | Integer | PK | - | A unique identifier of the group setting. |
| PaymentWallet_ID | Integer | FK -> mt5_pay_wallets.PaymentWallet_ID | - | Wallet configuration ID to which the group settings apply. |
| Group | String | - | - | The name of the group for which the wallet is available. |

## mt5_pay_wallet_countries

Country availability list for a wallet configuration.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Commission_ID | Integer | PK | - | A unique identifier of the country setting. (Column is named Commission_ID in the MT5 schema despite holding a country setting ID.) |
| PaymentWallet_ID | Integer | FK -> mt5_pay_wallets.PaymentWallet_ID | - | Wallet configuration ID to which the country settings apply. |
| Country | String | - | - | Two-letter names of the countries for which the wallet is available. |

## mt5_pay_wallet_commissions

Tiered commission rules for a wallet configuration (range-banded, per transaction type).

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Commission_ID | Integer | PK | - | A unique identifier of the commission setting. |
| PaymentWallet_ID | Integer | PK / FK -> mt5_pay_wallets.PaymentWallet_ID | - | Wallet configuration this commission belongs to (documented as part of the primary key). |
| Mode | Integer | - | EnCommissionMode | Commission calculation mode. |
| Entry | Integer | - | EnCommissionMode | Commission calculation mode depending on the transaction type. |
| Value | Integer | - | - | Commission sum. Commission units depend on the commission Mode. |
| RangeFrom | Float | - | - | Minimum transaction amount for which this commission will be charged. |
| RangeTo | Float | - | - | Maximum transaction amount for which this commission will be charged. |
| Minimal | Float | - | - | Minimum amount of commission charged. Specified in the group deposit currency. |
| Maximal | Float | - | - | Maximum amount of commission charged. Specified in the group deposit currency. |
| Currency | String | - | - | Commission calculation currency. Used for the COMM_MONEY_SPECIFIED mode. |

## mt5_pay_rules

Payment processing rules: a named, ordered rule that applies an action to transactions matching its conditions.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| PaymentRule_ID | Integer | PK | - | Unique identifier of the rule. |
| Timestamp | Integer | - | - | Unique value within the table; used internally. A changed Timestamp means the record was modified. |
| Name | String | - | - | Rule name. |
| Mode | Integer (type not stated in docs) | - | inline | Rule status: 0 — disabled, 1 — enabled. |
| Payment | (type not stated in docs) | - | - | Currently unused. |
| Flags | (type not stated in docs) | - | - | Currently unused. |
| Action | Integer (type not stated in docs) | - | EnAction | The action to be applied to the transaction that matches the rule conditions. |
| ActionDesc | String (type not stated in docs) | - | - | Rule description. |

## mt5_pay_rule_conditions

Individual match conditions belonging to a payment processing rule.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Condition_ID | Integer | PK | - | A unique identifier of the condition. |
| PaymentRule_ID | Integer | FK -> mt5_pay_rules.PaymentRule_ID | - | Identifier of the rule to which the condition applies. |
| Condition | Integer | - | EnCondition | Condition type. |
| Rule | Integer | - | EnConditionRule | The method used to compare the condition with the specified value. |
| ValueString | String | - | - | A string value for the condition. For example, for the "Symbols" condition the field holds symbol or symbol group names. |
| ValueInt | Integer | - | - | An int value for the condition. |
| ValueUInt | Integer | - | - | A uint value for the condition. |
| ValueFloat | Float | - | - | A float value for the condition. |

## mt5_corporate

Corporate link configurations: sets of branded links shown in terminals, scoped by account group, White Label company, and country.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Corporate_ID | Integer | PK | - | A unique configuration identifier, assigned automatically upon export. |
| Timestamp | Integer | - | - | Unique value within the table; used internally. A changed Timestamp means the record was modified. |
| Group | String | - | - | The group of accounts for which the links will be displayed. |
| Description | String | - | - | The description to be displayed in the list of settings. |
| Company | String | - | - | Company name (White Label) in whose terminals the links will be displayed. |
| Flags | Integer | - | - | Additional settings. Currently not in use. |
| Countries | Integer | - | - | The list of countries for which the links will be displayed. |

## mt5_corporate_links

The individual links belonging to a corporate link configuration.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Link_ID | Integer | PK | - | A unique link identifier, assigned automatically upon export. |
| Corporate_ID | Integer | FK -> mt5_corporate.Corporate_ID | - | Identifier of the configuration to which the link belongs. |
| CaptionType | Integer | - | EnCaptionType | Link type. |
| CaptionCustom | String | - | - | Custom link name. |
| Url | String | - | - | Link address. |
| Flags | Integer | - | - | Additional settings. Currently not in use. |

## mt5_network

General settings and live health/performance telemetry for every server in the platform. `Login` (server ID) is the key that all `mt5_network_*` and Anti-DDoS child tables join on.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Login | Integer | PK (server ID) | - | Server ID. |
| Timestamp | Integer | - | - | Unique value within the table; used internally. A changed Timestamp means the record was changed. |
| Type | Integer | - | - | Server type. |
| Name | String | - | - | Server name. |
| Address | String | - | - | Server address. |
| Port | String | - | - | Server port. |
| Adapter | String | - | - | Name of the currently used network controller. |
| ServiceTime | Integer | - | - | Service time (optimization window) when performance/reliability operations run. Minutes from 00:00 (e.g. 600 = 10:00). |
| FailoverMode | Integer | - | inline | Automatic failover mode: 0 — failover disabled; 1 — server is unavailable for most access servers; 2 — server is unavailable for all access servers. |
| FailoverTimeout | Integer | - | - | Seconds the server must be unavailable before monitoring servers start switching to the backup server. |
| Adapters | String | - | - | List of all available network controllers on the PC (comma-separated). |
| Addresses | String | - | - | List of available addresses for outgoing connections from this server (comma-separated). |
| Binds | String | - | - | List of listen addresses (comma-separated). |
| Points | String | - | - | List of public access points via which connections are accepted. |
| Version | Integer | - | - | Server version. |
| Build | Integer | - | - | Server build. |
| BuildDate | String | - | - | Server build date. |
| SysConnection | Integer | - | - | Status of the server's connection to the main trade server. |
| SysLastBoot | DateTime | - | - | Time of the last server boot (YYYY-MM-DD HH:MM:SS.MSC). |
| SysOsName | String | - | - | Operating system of the computer running the server. |
| SysCpuName | String | - | - | Processor type of the computer running the server. |
| SysCpuNumber | Integer | - | - | Number of CPU cores. |
| SysBits | Integer | - | inline | Operating system bitness: 32 — 32 bits; 64 — 64 bits; 0 — other. |
| SysMemoryTotal | Integer | - | - | Total amount of RAM in megabytes. |
| SysMemoryFree | Integer | - | - | Amount of free memory in megabytes. |
| SysMemoryCritical | Integer | - | - | Critical amount of free memory in megabytes. |
| SysHddSize | Integer | - | - | Total volume of the disk in megabytes. |
| SysHddFree | Integer | - | - | Free space on the disk in megabytes. |
| SysHddCritical | Integer | - | - | Critical amount of free space on the disk in megabytes. |
| SysHddFragmentation | Integer | - | - | Current fragmentation level of the server files, in percent. |
| SysHddFragCritical | Integer | - | - | Critical fragmentation level of the server files, in percent. |
| SysDefragRecommend | Integer | - | inline | Flag indicating that the OS recommends defragmenting the disk: 0 — no recommendation; 1 — recommendation present. |
| SysHddReadSpeed | Integer | - | - | Current disk read speed in MB/s. |
| SysHddReadCritical | Integer | - | - | Critical disk read speed in MB/s. |
| SysHddWriteSpeed | Integer | - | - | Current disk write speed in MB/s. |
| SysHddWriteCritical | Integer | - | - | Critical disk write speed in MB/s. |
| PerfConnectsMax | Integer | - | - | Maximum number of simultaneous connections to the server achieved during the day. |
| PerfConnectsCritical | Integer | - | - | Critical number of simultaneous connections to the server. |
| PerfCpuMax | Integer | - | - | Maximum level of CPU usage in percent for the current day. |
| PerfCpuCritical | Integer | - | - | Critical level of CPU usage in percent. |
| PerfMemoryMin | Integer | - | - | Minimum size of free RAM and virtual memory in megabytes per day. |
| PerfMemoryCritical | Integer | - | - | Critical amount of free memory in megabytes. |
| PerfMemBlockMin | Integer | - | - | Minimum value of the maximum memory block in megabytes per day. |
| PerfMemBlockCritical | Integer | - | - | Critical value of the maximum memory block in megabytes per day. |
| PerfNetworkMax | Integer | - | - | Maximum level of network usage in KB/s for the current day. |
| PerfNetworkCritical | Integer | - | - | Critical level of network usage in KB/s. |
| PerfSocketsMax | Integer | - | - | Maximum number of active sockets per day. |
| PerfSocketsCritial | Integer | - | - | Critical number of active sockets. (Field name is misspelled in the MT5 schema.) |

## mt5_network_trade_servers

Trade-server-specific settings: demo account policy, overnight/overmonth rollover, ticket and login ranges, and aggregate counters.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Login | Integer | PK / FK -> mt5_network.Login | - | Server ID. |
| DemoMode | Integer | - | inline | Mode of demo account allocation: 0 — creation of demo accounts disabled; 1 — prolong the period of demo accounts after connection; 2 — demo accounts with a fixed expiration date. |
| DemoPeriod | Integer | - | - | The validity period of demo accounts. |
| OvernightMode | Integer | - | - | The overnight mode. |
| OvernightTime | Integer | - | - | The time of transition to the next day, in minutes after 00:00. |
| OvernightTimeLast | DateTime | - | - | Time of the last transition to the next day (YYYY-MM-DD HH:MM:SS.MSC). |
| OvernightTimePrev | DateTime | - | - | Time of the penultimate transition to the next day (YYYY-MM-DD HH:MM:SS.MSC). |
| OvernightDays | Integer | - | inline bitmask | Schedule of trading-day-closure operations, as a sum of flags: 0x01 Sun, 0x02 Mon, 0x04 Tue, 0x08 Wed, 0x10 Thu, 0x20 Fri, 0x40 Sat. Further flags specify days on which swaps are charged: 0x80 Sun, 0x100 Mon, 0x200 Tue, 0x400 Wed, 0x800 Thu, 0x1000 Fri, 0x2000 Sat. |
| OvermonthMode | Integer | - | inline | The overmonth mode: 0 — on the last day of the month; 1 — on the first day of the month. |
| OvermonthTimeLast | DateTime | - | - | Time of the last transition to the next month (YYYY-MM-DD HH:MM:SS.MSC). |
| OvermonthTimePrev | DateTime | - | - | Time of the penultimate transition to the next month (YYYY-MM-DD HH:MM:SS.MSC). |
| TotalUsers | Integer | - | - | Total number of client accounts on the trade server. |
| TotalUsersReal | Integer | - | - | Total number of real clients on the trade server. |
| TotalDeals | Integer | - | - | Total number of deals executed on the trade server. |
| TotalOrders | Integer | - | - | Total number of active orders placed on the trade server. |
| TotalOrdersHistory | Integer | - | - | Total number of orders in the history on the trade server. |
| TotalPositions | Integer | - | - | Total number of positions on the trade server. |
| LoginsRange | String | - | - | Account ranges on the trade server, e.g. 1000-1000000,2000001-2999001. |
| LoginsRangeUsed | String | - | - | Ranges of actually used logins from LoginsRange, e.g. 1000-1004,0-0. |
| OrdersRange | String | - | - | Order ticket ranges on the trade server, e.g. 1-1000000,2000001-2999001. |
| OrdersRangeUsed | String | - | - | Ranges of actually used order tickets from OrdersRange, e.g. 1-53,0-0. |
| DealsRange | String | - | - | Trade ticket ranges on the trade server, e.g. 1-1000000,2000001-2999001. |
| DealsRangeUsed | String | - | - | Ranges of actually used trade tickets from DealsRange, e.g. 1-56,0-0. |

## mt5_network_history_servers

History-server-specific settings.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Login | Integer | PK / FK -> mt5_network.Login | - | Server ID. |
| DatafeedsTimeout | Integer | - | - | Timeout of data feeds before switching to other ones, in seconds. |
| NewsMax | Integer | - | - | Maximum number of news items that can be stored on the history server. |

## mt5_network_backup_servers

Backup-server-specific settings, including the SQL export configuration that produces every table in this catalog.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Login | Integer | PK / FK -> mt5_network.Login | - | Server ID. |
| PairrServer | Integer | FK -> mt5_network.Login | - | ID of the server to back up. (Field name is misspelled in the MT5 schema.) |
| BackupFlags | Integer | - | inline bitmask | Backup settings as a sum of flags: 1 — enables backup; 2 — enables backup of tick data; 4 — allows the server to be used for automatic failover; 8 — enables synchronization of logs with the primary server. |
| BackupPath | String | - | - | The path to save backups. |
| BackupPeriod | Integer | - | inline | Backup frequency: 0 — no periodic backups; 1 — every 15 minutes; 2 — every 30 minutes; 3 — every hour; 4 — every 4 hours; 5 — every day. |
| BackupTtl | Integer | - | inline | Period to keep backups: 1 — one day; 2 — three days; 3 — one week; 4 — one month; 5 — three months; 6 — six months. |
| BackupTimeFull | Integer | - | - | Time of creating full backup copies, in minutes since 00:00. |
| BackupLastStartup | DateTime | - | - | Last backup copy creation time at server launch (YYYY-MM-DD HH:MM:SS.MSC). |
| BackupLastFull | DateTime | - | - | Last full backup copy creation time (YYYY-MM-DD HH:MM:SS.MSC). |
| BackupLastArchive | DateTime | - | - | Last incremental backup copy creation time (YYYY-MM-DD HH:MM:SS.MSC). |
| BackupLastSync | DateTime | - | - | Time of the last successful synchronization with the backed-up server (YYYY-MM-DD HH:MM:SS.MSC). |
| SqlMode | Integer | - | inline | Mode of exporting data to an SQL database: 0 — export disabled; 1 — Microsoft SQL Server; 2 — FireBird; 3 — MySQL; 4 — Oracle. |
| SqlServer | String | - | - | Address of the server the database is installed on. |
| SqlFolder | String | - | - | Name of the SQL database the data is exported to. |
| SqlFlags | Integer | - | inline bitmask | Additional SQL export settings, sum of flags: 1 — export trade history to separate tables; 2 — do not export accounts and trade operations of demo groups. E.g. 3 means both enabled. |
| SqlPeriod | Integer | - | - | Frequency of price and profit data export. |
| SQLExportLastSync | Integer | - | - | Time of the last full synchronization of databases and platform configurations with the SQL database, in seconds since 01.01.1970. 0 if export is disabled or synchronization is in progress. |

## mt5_network_backup_folders

Custom directories included in a backup server's backup set.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Folder_ID | Integer | PK | - | Unique entry ID. |
| Login | Integer | FK -> mt5_network.Login (backup server) | - | Backup server ID. |
| Folder | String | - | - | Path to the backed-up folder, relative to the backup server installation directory. |
| Masks | String | - | - | List of copied files (comma-separated). Files can be specified by masks. |
| Filter | String | - | - | List of ignored files (comma-separated). Files can be specified by masks. |

## mt5_network_access_servers

Access-server-specific settings: priority, antiflood limits, load-balancing state, and permitted connection types.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Login | Integer | PK / FK -> mt5_network.Login | - | Server ID. |
| Priority | Integer | - | - | Base priority of the access server, 0-15. The special priority 255 (idle) creates a backup access server. |
| AntifloodEnable | Integer | - | inline | Antiflood control: 0 — disabled, 1 — enabled. |
| AntifloodConnects | Integer | - | - | Maximum number of connections from one IP address within a period, after which the address is temporarily blocked. |
| AntifloodErrors | Integer | - | - | Maximum number of incorrect connections after which the IP address is temporarily blocked. |
| NewsMaxCount | Integer | - | - | Maximum number of news items that can be stored on the access server. |
| BalancingConnections | Integer | - | - | Current number of connections. |
| BalancingPriority | Integer | - | - | Current priority of the access server. |
| AccessMask | Integer | - | inline bitmask (cf. EnAccessMask) | Allowed connection types, as a sum of flags: 1 — client; 2 — manager; 4 — administrator; 8 — Client API; 16 — Manager API; 32 — Web API. E.g. 63 means all types allowed. |
| AccessFlags | Integer | - | inline | Additional server access flags: 1 — the access server is hidden from all terminals but is available for connection. |
| Servers | String | - | - | IDs of trade servers (comma-separated) reachable through this access server. |

## mt5_network_antiddos

Anti-DDoS proxy server settings.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Login | Integer | PK / FK -> mt5_network.Login | - | Server ID. |
| Priority | Integer | - | - | Base server priority, 0-15. The special priority 255 (idle) creates a backup Anti-DDoS server. |
| AccessMask | Integer | - | EnAccessMask | Allowed types of connection to the server. |

## mt5_antiddos_servers

Trading servers reachable through a given Anti-DDoS proxy server.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Server_ID | Integer | PK | - | A unique configuration identifier, assigned automatically upon export. |
| Login | Integer | FK -> mt5_network_antiddos.Login | - | Identifier of the Anti-DDoS server to which the trading server belongs. |
| Server | Integer | FK -> mt5_network.Login | - | Identifier of the trading server accessed through this Anti-DDoS server. |

## mt5_antiddos_sources

IP address ranges belonging to the Anti-DDoS provider's proxy servers (source allowlist).

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Source_ID | Integer | PK | - | A unique configuration identifier, assigned automatically upon export. |
| Login | Integer | FK -> mt5_network_antiddos.Login | - | Identifier of the Anti-DDoS server to which the list belongs. |
| From | String | - | - | Start of IP address range. |
| To | String | - | - | End of IP address range. |

## mt5_firewall

Platform firewall rules (IP-range allow/block list). Note: no surrogate primary key is documented — `Rule_Index` carries the ordering.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Timestamp | Integer | - | - | Unique value within the table; used internally. A changed Timestamp means the record was changed. |
| Action | Integer | - | inline | Type of action taken by the firewall rule: 0 — block; 1 — allow; 2 — always allow. |
| From | String | - | - | Beginning of the IP address range the rule applies to. |
| To | String | - | - | End of the IP address range the rule applies to. |
| Comment | String | - | - | A comment to the firewall rule. |
| Rule_Index | Integer | ordering key | - | The index number of the configuration in the list, starting from 0. |

## mt5_messengers

Messenger and SMS provider configurations used for outbound notifications (2FA codes, alerts).

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Messenger_ID | Integer | PK | - | A unique identifier of the configuration. |
| Timestamp | Integer | - | - | Unique value within the table; used internally. A changed Timestamp means the record was modified. |
| Name | String | - | - | Name of the messenger. |
| Sender | String | - | - | Sender name in the messenger configuration. |
| ProviderType | Integer | - | EnProviderType | Type of message service provider in the messenger configuration. |
| ProviderAddress | String | - | - | Server address of the provider in the messenger configuration. |
| ProviderLogin | String | - | - | Login of the account used to send messages via the messenger. |
| ProviderPassword | String | - | - | Password of the account used to send messages via the messenger. |
| ProviderToken | String | - | - | Authentication token used to send messages via the messenger. |
| ProviderSubId | String | - | - | Sender ID used to send messages via the messenger. |
| ProviderCurrency | String | - | - | Currency in which the provider's services are billed. |
| ProviderCurrecnyRate | Float | - | - | Conversion rate of the provider's service currency to USD. (Field name is misspelled in the MT5 schema.) |
| Flags | Integer | - | EnFlags | Additional messenger settings. |
| MessageFormat | String | - | - | Template used to send messages through this messenger. |

## mt5_messenger_groups

Per-account-group overrides for a messenger configuration.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Group_ID | Integer | PK | - | A unique identifier of the setting. |
| Messenger_ID | Integer | FK -> mt5_messengers.Messenger_ID | - | Identifier of the configuration associated with this setting. |
| Group | String | - | - | Full path to the group. |
| Sender | String | - | - | Message sender name. |

## mt5_messenger_countries

Per-country overrides for a messenger configuration.

| Field | Type | PK/FK | Enum | Description |
|---|---|---|---|---|
| Country_ID | Integer | PK | - | A unique identifier of the setting. |
| Messenger_ID | Integer | FK -> mt5_messengers.Messenger_ID | - | Identifier of the configuration associated with this setting. |
| PhoneCode | Integer | - | - | Country phone code. |
| MessageFormat | String | - | - | Message template used for this country. |

---

## Foreign-key map

```
mt5_pay_wallets (PaymentWallet_ID)
  <- mt5_pay_wallet_params.PaymentWallet_ID
  <- mt5_pay_wallet_groups.PaymentWallet_ID
  <- mt5_pay_wallet_countries.PaymentWallet_ID
  <- mt5_pay_wallet_commissions.PaymentWallet_ID
mt5_pay_wallets.Server            -> mt5_network.Login

mt5_pay_rules (PaymentRule_ID)
  <- mt5_pay_rule_conditions.PaymentRule_ID

mt5_corporate (Corporate_ID)
  <- mt5_corporate_links.Corporate_ID

mt5_network (Login)
  <- mt5_network_trade_servers.Login
  <- mt5_network_history_servers.Login
  <- mt5_network_backup_servers.Login  (and .PairrServer -> mt5_network.Login)
  <- mt5_network_backup_folders.Login
  <- mt5_network_access_servers.Login
  <- mt5_network_antiddos.Login

mt5_network_antiddos (Login)
  <- mt5_antiddos_servers.Login   (mt5_antiddos_servers.Server -> mt5_network.Login)
  <- mt5_antiddos_sources.Login

mt5_messengers (Messenger_ID)
  <- mt5_messenger_groups.Messenger_ID
  <- mt5_messenger_countries.Messenger_ID

mt5_payments / mt5_payment_history
  .Login, .Manager -> mt5_users.Login
  .DealID          -> mt5_deals.Deal

mt5_firewall — standalone, no FK (ordered by Rule_Index)
```

## Enumerations referenced by this domain

`EnPaymentAction`, `EnPaymentState`, `EnPaymentType`, `EnPaymentFlags`,
`EnPaymentProviderActionFlags`, `EnUsersConnectionTypes`, `EnCommissionMode`,
`EnAction`, `EnCondition`, `EnConditionRule`, `EnCaptionType`, `EnAccessMask`,
`EnProviderType`, `EnFlags`.

Several other columns use undocumented inline value lists rather than a named enumeration
(marked `inline` / `inline bitmask` above): wallet `Enable`, `pay_wallet_params.Type`,
`pay_rules.Mode`, `network.FailoverMode` / `SysBits` / `SysDefragRecommend`,
`trade_servers.DemoMode` / `OvernightDays` / `OvermonthMode`,
`backup_servers.BackupFlags` / `BackupPeriod` / `BackupTtl` / `SqlMode` / `SqlFlags`,
`access_servers.AntifloodEnable` / `AccessMask` / `AccessFlags`, and `firewall.Action`.
