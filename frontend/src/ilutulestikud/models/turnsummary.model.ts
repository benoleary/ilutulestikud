export class TurnSummary
{
    GameIdentifier: string;
    GameName: string;
    IsPlayerTurn: boolean;

    constructor(turnSummaryObject: Object)
    {
        this.refreshFromSource(turnSummaryObject);
    }

    refreshFromSource(turnSummaryObject: Object)
    {
        this.GameIdentifier = turnSummaryObject["GameIdentifier"];
        this.GameName = turnSummaryObject["GameName"];
        this.IsPlayerTurn = turnSummaryObject["IsPlayerTurn"];
    }
}
