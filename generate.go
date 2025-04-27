package ordersystem

//go:generate cp --archive --dereference --target-directory ./html ../websites/order.proxysto.re
//go:generate gotext-update-templates -srclang=en-US -lang=de-DE -out=catalog.go . ./html
