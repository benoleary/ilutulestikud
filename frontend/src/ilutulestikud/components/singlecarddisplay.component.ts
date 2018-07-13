import { Component, Input } from '@angular/core';
import { VisibleCard } from '../models/visiblecard.model';


@Component({
    selector: 'single-card-display',
    templateUrl: './singlecarddisplay.component.html',
  })
  export class SingleCardDisplayComponent
  {
    @Input() singleCard: VisibleCard;

    constructor()
    {
        this.singleCard = null;
    }
  }