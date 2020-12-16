# hermezon
Hermezon is a service that allows you to notify a phone number either when a product in Amazon becomes available or when its price drops below some target price.


## To Do

- Refactor database access to use interfaces so we could switch the implementation in the future
- Refactor availability.Run and price.Run actions
- Ability to send messages to multipler "messengers": Currently we send messages either to Twilio or to Telegram, but not both.
- Ability to send telegram chat id in the "from" field of a request to the API. -> This should be done, but we need to refactor the "phone" field to another naming, because in the case of telegram it's not a "phone", but a channel ID.
- Add a "channel" field to set whether we want to use Telegram or Twilio.
- Add tests
