I found the way to write in private channels.

    You should convert it to public with some @channelName

    Send a message to this channel through the Bot API:

    https://api.telegram.org/bot111:222/sendMessage?chat_id=@channelName&text=123

    As the response, you will get information with chat_id of your channel.
<pre>
    {
      "ok" : true,
      "result" : {
        "chat" : {
          **"id" : -1001005582487,**
          "title" : "Test Private Channel",
          "type" : "channel"
        },
        "date" : 1448245538,
        "message_id" : 7,
        "text" : "123ds"
      }
    }
</pre>

    Now you can convert the channel back to private (by deleting the channel's link) and send a message directly to the chat_id "-1001005582487":

    https://api.telegram.org/bot111:222/sendMessage?chat_id=-1001005582487&text=123

