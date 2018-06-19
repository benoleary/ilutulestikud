export class VisibleCard
{
    ColorSuit: string;
    SequenceIndex: number;

    constructor(cardObject: Object)
    {
        this.RefreshFromSource(cardObject);
    }

    RefreshFromSource(cardObject: Object)
    {
        this.ColorSuit = cardObject["ColorSuit"];
        this.SequenceIndex = cardObject["SequenceIndex"];

        // This is a temporary solution until we display images rather than colored text,
        // as there is no "rainbow" color.
        if (this.ColorSuit == "rainbow")
        {
            this.ColorSuit = "pink";
        }
    }

    static RefreshListFromSource(listToRefresh: VisibleCard[], cardObjectList: Object[])
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
                listToRefresh.push(new VisibleCard(fetchedCard));
            }
        }
    }

    static RefreshListOfListsFromSource(listOfListsToRefresh: VisibleCard[][], cardObjectListOfLists: Object[][])
    {
        // First of all we reduce the number of cards if there were more than request gave us.
        if (listOfListsToRefresh.length > cardObjectListOfLists.length)
        {
            listOfListsToRefresh.length = cardObjectListOfLists.length;
        }

        // Next we make sure that there are empty arrays if we do not have enough elements in the outer array.
        while (listOfListsToRefresh.length < cardObjectListOfLists.length)
        {
            listOfListsToRefresh.push([]);
        }

        for (var listIndex: number = 0; listIndex < cardObjectListOfLists.length; ++listIndex)
        {
            VisibleCard.RefreshListFromSource(listOfListsToRefresh[listIndex], Array.from(cardObjectListOfLists[listIndex]))
        }
    }
}