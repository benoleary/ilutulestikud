import { Component } from '@angular/core';
import { Inject } from '@angular/core';
import { MatDialogRef } from '@angular/material';
import { MAT_DIALOG_DATA } from '@angular/material';
import { VisibleCard } from '../models/visiblecard.model';
import { InferredCard } from '../models/inferredcard.model';

@Component({
    selector: 'hand-details-display',
    templateUrl: './handdetailsdisplay.component.html',
  })
  export class HandDetailsDisplayComponent
  {
    cardsWithPossibilities: [VisibleCard, InferredCard][];

    constructor(
        public dialogReference: MatDialogRef<HandDetailsDisplayComponent>,
        @Inject(MAT_DIALOG_DATA) public data: any)
    {
        this.cardsWithPossibilities = [];

        if (data && data.visibleCards && data.inferredCards)
        {
            const numberOfCards: number = data.visibleCards.length
            for (let cardIndex: number = 0; cardIndex < numberOfCards; cardIndex += 1)
            {
                if (cardIndex >= data.inferredCards.length)
                {
                    console.log("lengths of card arrays did not match: data = " + JSON.stringify(data));
                    continue;
                }

                this.cardsWithPossibilities.push(
                    [data.visibleCards[cardIndex], data.inferredCards[cardIndex]])
            }

        }
    }

    closeDialog(): void
    {
        this.dialogReference.close(null);
    }
  }