import { VisibleCard } from './visiblecard.model';

export class VisibleHand
{
    playerName: string;
    visibleCards: VisibleCard[]

    constructor(handObject: Object)
    {
        this.refreshFromSource(handObject);
    }

    refreshFromSource(handObject: Object)
    {
        console.log("refreshFromSource(handObject: " + JSON.stringify(handObject) + ")")

        this.playerName = handObject["PlayerName"];
        this.visibleCards = Array.from(handObject["HandCards"]);
    }

    static refreshListFromSource(listToRefresh: VisibleHand[], handObjectList: Object[])
    {
        console.log("refreshListFromSource(listToRefresh: " + JSON.stringify(listToRefresh) 
        + ", handObjectList: " + JSON.stringify(handObjectList) + ")")

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
                listToRefresh[handIndex].refreshFromSource(fetchedHand);
            }
            else
            {
                listToRefresh.push(new VisibleHand(fetchedHand));
            }
        }
    }
}