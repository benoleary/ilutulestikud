import { Component, Input } from '@angular/core';
import { VisibleCard } from '../models/visiblecard.model';


@Component({
    selector: 'card-array-display',
    templateUrl: './cardarraydisplay.component.html',
  })
  export class CardArrayDisplayComponent
  {
    @Input() cardArray: VisibleCard[];

    constructor()
    {
        this.cardArray = null;
    }

    hasCardsToDisplay(): boolean
    {
      return this.cardArray && (this.cardArray.length > 0);
    }
  }