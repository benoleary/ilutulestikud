export class TurnSummary
{
    GameName: string;
    CreationTimestampInSeconds: number;
    TurnNumber: number;
    PlayersInNextTurnOrder: string[];
    IsPlayerTurn: boolean;

    constructor(turnSummaryObject: Object)
    {
        this.GameName = turnSummaryObject["GameName"];
        this.CreationTimestampInSeconds = turnSummaryObject["CreationTimestampInSeconds"];
        this.TurnNumber = turnSummaryObject["TurnNumber"];
        this.PlayersInNextTurnOrder = turnSummaryObject["PlayersInNextTurnOrder"];
        this.IsPlayerTurn = turnSummaryObject["IsPlayerTurn"];
    }
}
