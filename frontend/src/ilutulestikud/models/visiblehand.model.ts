import { VisibleCard } from './visiblecard.model';
import { InferredCard } from './inferredcard.model';

export class VisibleHand
{
    playerName: string;
    playerColor: string;
    visibleCards: VisibleCard[]
    knowledgeOfCards: InferredCard[]

    constructor(handObject: Object)
    {
        this.RefreshFromSource(handObject);
    }

    RefreshFromSource(handObject: Object)
    {
        this.playerName = handObject["PlayerName"];
        this.playerColor = handObject["PlayerColor"];

        // Simply making an array out of the parsed array-like object does not call any constructors.
        this.visibleCards = [];
        Array.from(handObject["HandCards"])
             .forEach(cardAsObject => this.visibleCards.push(new VisibleCard(cardAsObject)));
        this.knowledgeOfCards = [];
        Array.from(handObject["KnowledgeOfOwnHand"])
             .forEach(cardAsObject => this.knowledgeOfCards.push(new InferredCard(cardAsObject)));
    }

    static RefreshListFromSource(listToRefresh: VisibleHand[], handObjectList: Object[])
    {
        // First of all we reduce the number of hands if there were more than the request gave us.
        if (listToRefresh.length > handObjectList.length)
        {
            listToRefresh.length = handObjectList.length;
        }

        for (var handIndex: number = 0; handIndex < handObjectList.length; ++handIndex)
        {
            const fetchedHand: Object = handObjectList[handIndex];

            // We could replace each hand with each refresh, but to avoid possible issues (such
            // as happens with the turn summaries), we update existing hands and only add new
            // ones when necessary.
            if (handIndex < listToRefresh.length)
            {
                listToRefresh[handIndex].RefreshFromSource(fetchedHand);
            }
            else
            {
                listToRefresh.push(new VisibleHand(fetchedHand));
            }
        }
    }
}