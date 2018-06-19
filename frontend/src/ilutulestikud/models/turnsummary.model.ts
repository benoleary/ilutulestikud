import { BackendIdentification } from './backendidentification.model';

export class TurnSummary
{
    GameForBackend: BackendIdentification;
    IsPlayerTurn: boolean;

    constructor(turnSummaryObject: Object)
    {
        this.RefreshFromSource(turnSummaryObject);
    }

    RefreshFromSource(turnSummaryObject: Object)
    {
        this.GameForBackend
          = new BackendIdentification(
              turnSummaryObject["GameName"],
              turnSummaryObject["GameIdentifier"])
        this.IsPlayerTurn = turnSummaryObject["IsPlayerTurn"];
    }

    GameName(): string
    {
        return this.GameForBackend.NameForPost
    }
}
