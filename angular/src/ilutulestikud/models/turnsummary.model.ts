export class TurnSummary
{
    GameName: string;
    CreationTimestampInSeconds: number;
    TurnNumber: number;
    PlayersInNextTurnOrder: string[];
    IsPlayerTurn: boolean;

    constructor(turnSummaryObject: Object)
    {
        this.refreshFromSource(turnSummaryObject);
    }

    refreshFromSource(turnSummaryObject: Object)
    {
        this.GameName = turnSummaryObject["GameName"];
        this.CreationTimestampInSeconds = turnSummaryObject["CreationTimestampInSeconds"];
        this.TurnNumber = turnSummaryObject["TurnNumber"];
        this.PlayersInNextTurnOrder = turnSummaryObject["PlayersInNextTurnOrder"];
        this.IsPlayerTurn = turnSummaryObject["IsPlayerTurn"];
    }

    asText(): string
    {
        // Date constructors take a timestamp in units of milliseconds.
        return "Created: " + (new Date(this.CreationTimestampInSeconds * 1000)).toTimeString()
             + "; turn number: " + this.TurnNumber
             + "; player order: " + this.PlayersInNextTurnOrder.join(", ");
    }
}
