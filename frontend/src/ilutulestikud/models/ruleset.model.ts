export class Ruleset
{
    Identifier: string;
    Description: string;
    MinimumNumberOfPlayers: number;
    MaximumNumberOfPlayers: number;

    constructor(rulesetObject: Object)
    {
        this.Identifier = rulesetObject["Identifier"];
        this.Description = rulesetObject["Description"];
        this.MinimumNumberOfPlayers = rulesetObject["MinimumNumberOfPlayers"];
        this.MaximumNumberOfPlayers = rulesetObject["MaximumNumberOfPlayers"];
    }
}