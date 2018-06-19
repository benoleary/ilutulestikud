export class InferredCard
{
    PossibleColorSuits: string[];
    PossibleSequenceIndices: number[];

    constructor(cardObject: Object)
    {
        this.RefreshFromSource(cardObject);
    }

    RefreshFromSource(cardObject: Object)
    {
        this.PossibleColorSuits = cardObject["PossibleColorSuits"];
        this.PossibleSequenceIndices = cardObject["PossibleSequenceIndices"];
    }

    static RefreshListFromSource(listToRefresh: InferredCard[], cardObjectList: Object[])
    {
        // First of all we reduce the number of cards if there were more than the request gave us.
        if (listToRefresh.length > cardObjectList.length)
        {
            listToRefresh.length = cardObjectList.length;
        }

        for (var cardIndex: number = 0; cardIndex < cardObjectList.length; ++cardIndex)
        {
            const fetchedCard: Object = cardObjectList[cardIndex];

            // We could replace each card with each refresh, but to avoid possible issues (such
            // as happens with the turn summaries), we update existing cards and only add new
            // ones when necessary.
            if (cardIndex < listToRefresh.length)
            {
                listToRefresh[cardIndex].RefreshFromSource(fetchedCard)
            }
            else
            {
                listToRefresh.push(new InferredCard(fetchedCard));
            }
        }
    }
}