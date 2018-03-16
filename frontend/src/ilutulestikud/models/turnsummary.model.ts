export class TurnSummary
{
    GameIdentifier: string;
    GameName: string;
    CreationTimestampInSeconds: number;
    TurnNumber: number;
    PlayerNamesInNextTurnOrder: string[];
    IsPlayerTurn: boolean;

    constructor(turnSummaryObject: Object)
    {
        this.refreshFromSource(turnSummaryObject);
    }

    refreshFromSource(turnSummaryObject: Object)
    {
        this.GameIdentifier = turnSummaryObject["GameIdentifier"];
        this.GameName = turnSummaryObject["GameName"];
        this.CreationTimestampInSeconds = turnSummaryObject["CreationTimestampInSeconds"];
        this.TurnNumber = turnSummaryObject["TurnNumber"];
        this.PlayerNamesInNextTurnOrder = turnSummaryObject["PlayerNamesInNextTurnOrder"];
        this.IsPlayerTurn = turnSummaryObject["IsPlayerTurn"];
    }

    asText(): string
    {
        // Date constructors take a timestamp in units of milliseconds.
        return "Created: " + (new Date(this.CreationTimestampInSeconds * 1000)).toTimeString()
             + "; turn number: " + this.TurnNumber
             + "; player order: " + this.PlayerNamesInNextTurnOrder.join(", ");
    }
}
