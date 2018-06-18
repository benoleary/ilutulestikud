export class LogMessage
{
    Color: string;
    Text: string;

    constructor(messageObject: Object)
    {
        this.refreshFromSource(messageObject);
    }

    refreshFromSource(messageObject: Object)
    {
        this.Color = messageObject["TextColor"];
        const timestampInSeconds = messageObject["TimestampInSeconds"];
        const playerName = messageObject["PlayerName"];
        if (!playerName)
        {
            this.Text = "";
        }
        else
        {
            this.Text = (new Date(timestampInSeconds * 1000)).toTimeString()
                + " - " + messageObject["PlayerName"]
                + ": " + messageObject["MessageText"];
        }
    }

    static refreshListFromSource(listToRefresh: LogMessage[], messageObjectList: Object[])
    {
        // First of all we reduce the number of log lines if there were more than the request gave us.
        if (listToRefresh.length > messageObjectList.length)
        {
            listToRefresh.length = messageObjectList.length;
        }

        for (var messageIndex: number = 0; messageIndex < messageObjectList.length; ++messageIndex)
        {
            const fetchedMessage: Object = messageObjectList[messageIndex];

            // We could replace each message with each refresh, but to avoid possible issues (such
            // as happens with the turn summaries), we update existing messages and only add new ones
            // when necessary.
            if (messageIndex < listToRefresh.length)
            {
                listToRefresh[messageIndex].refreshFromSource(fetchedMessage);
            }
            else
            {
                listToRefresh.push(new LogMessage(fetchedMessage));
            }
        }
    }
}