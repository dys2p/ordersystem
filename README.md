# ordersystem

ordersystem uses an SQLite Database, continuous replication using [Litestream](https://litestream.io) is recommended.

## Please note

* Check your BTCPay Server regularly for invoices which have been paid partially or late

## BTCPay Server Configuration

* User API Keys: enable `btcpay.store.canviewinvoices` and `btcpay.store.cancreateinvoice`
* Store Webhook
  * Payload URL: `https://example.com/rpc`
  * Automatic redelivery: yes
  * Is enabled: yes
  * Events: "A new payment has been received", "An invoice has been settled"
* Store access token
  * PublicKey: use the hex SIN which ordersystem writes to the log on startup

## ToDo

* client should see task ID
* reshipping: integrate shipment number (if desired) into user interface
