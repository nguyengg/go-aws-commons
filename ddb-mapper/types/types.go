// Package types defines some utility classes to be used as Go type for DynamoDB attributes.
//
// [StringSet] is implemented internally as a map[string]struct{}. It will marshal to "SS" data type if non-empty. If
// empty, it will marshal to "NULL" data type. Its zero value is ready for use. If you want to mimic the behaviour of
// this class, you must not forget to tag the field as ",stringset" lest the encoder will use "L" data type instead.
//
// [UnixTime] and [UnixMilli] are [time.Time] aliases that marshal as unix time in seconds and in milliseconds ("N" data
// type) respectively.
package types
