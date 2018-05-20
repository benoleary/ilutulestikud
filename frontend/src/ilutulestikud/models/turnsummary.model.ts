import { BackendIdentification } from './backendidentification.model';

export class TurnSummary
{
    GameForBackend: BackendIdentification;
    IsPlayerTurn: boolean;

    constructor(turnSummaryObject: Object)
    {
        this.refreshFromSource(turnSummaryObject);
    }

    refreshFromSource(turnSummaryObject: Object)
    {
        this.GameForBackend
          = new BackendIdentification(
              turnSummaryObject["GameName"],
              turnSummaryObject["GameIdentifier"])
        this.IsPlayerTurn = turnSummaryObject["IsPlayerTurn"];
    }
}
