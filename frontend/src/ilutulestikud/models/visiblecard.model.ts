export class VisibleCard
{
    ColorSuit: string;
    SequenceIndex: number;

    constructor(cardObject: Object)
    {
        this.ColorSuit = cardObject["ColorSuit"];
        this.SequenceIndex = cardObject["SequenceIndex"];
    }
}