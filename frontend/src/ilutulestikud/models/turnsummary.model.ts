export class TurnSummary
{
    GameIdentifier: string;
    GameName: string;
	RulesetDescription: string;
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
        this.RulesetDescription = turnSummaryObject["RulesetDescription"];
        this.CreationTimestampInSeconds = turnSummaryObject["CreationTimestampInSeconds"];
        this.TurnNumber = turnSummaryObject["TurnNumber"];
        this.PlayerNamesInNextTurnOrder = turnSummaryObject["PlayerNamesInNextTurnOrder"];
        this.IsPlayerTurn = turnSummaryObject["IsPlayerTurn"];
    }

    asTextLines(): string[]
    {
        // Date constructors take a timestamp in units of milliseconds.
        return ["Created: " + (new Date(this.CreationTimestampInSeconds * 1000)).toTimeString(),
        "Ruleset: " + this.RulesetDescription,
        "Turn number: " + this.TurnNumber,
        "Player order: " + this.PlayerNamesInNextTurnOrder.join(", ")];
    }
}
