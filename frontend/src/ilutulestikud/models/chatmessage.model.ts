export class ChatMessage
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
}