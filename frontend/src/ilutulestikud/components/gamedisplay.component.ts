import { Component, Input, Output, EventEmitter } from '@angular/core';
import { OnInit, OnDestroy } from '@angular/core/src/metadata/lifecycle_hooks';
import { Observable } from 'rxjs/Rx';
import { Subscription } from 'rxjs/Subscription';
import { IlutulestikudService } from '../ilutulestikud.service';
import { LogMessage } from '../models/logmessage.model';
import { BackendIdentification } from '../models/backendidentification.model';
import { VisibleCard } from '../models/visiblecard.model';
import { VisibleHand } from '../models/visiblehand.model';


@Component({
    selector: 'game-display',
    templateUrl: './gamedisplay.component.html',
  })
  export class GameDisplayComponent implements OnInit, OnDestroy
  {
    @Input() gameIdentification: BackendIdentification;
    @Input() playerIdentification: BackendIdentification;
    @Output() onError: EventEmitter<any> = new EventEmitter();
    informationText: string;
    gameDataSubscription: Subscription;
    isAwaitingGameData: boolean;
    chatLog: LogMessage[];
    chatInput: string;
    actionLog: LogMessage[];
    noncardInformationText: string;
    playedSequences: VisibleCard[][];
    discardPile: VisibleCard[];
    handsBeforeViewingPlayer: VisibleHand[];
    handsAfterViewingPlayer: VisibleHand[];

    constructor(public ilutulestikudService: IlutulestikudService)
    {
        this.gameIdentification = null;
        this.playerIdentification = null;
        this.chatLog = [];
        this.chatInput = null;
        this.actionLog = [];
        this.noncardInformationText = null;
        this.playedSequences = [];
        this.discardPile = [];
        this.handsBeforeViewingPlayer = [];
        this.handsAfterViewingPlayer = [];
    }

    ngOnInit(): void
    {
        this.gameDataSubscription =
          Observable
            .timer(0, 1000)
            .takeWhile(() => (this.gameIdentification != null))
            .subscribe(
              () => this.refreshGameData(),
              thrownError => this.onError.emit(thrownError),
              () => {});
    }
  
    ngOnDestroy(): void
    {
      if (this.gameDataSubscription)
      {
        this.gameDataSubscription.unsubscribe();
      }
  
      this.gameIdentification = null;
    }

    refreshGameData(): void
    {
      // We only request new game data if we are not waiting for the response to the last request.
      if (!this.isAwaitingGameData)
      {
        // We note that we are now awaiting the HTTP response (this.isAwaitingGameData will be set
        // back to false by this.displayGameData(fetchedGameData) which will run when we get the
        // response to the request).
        this.isAwaitingGameData = true;
        this.ilutulestikudService
          .gameAsSeenByPlayer(this.gameIdentification, this.playerIdentification)
          .subscribe(
            fetchedGameData => this.parseGameData(fetchedGameData),
            thrownError => this.onError.emit(thrownError),
            () => {});
      }
    }

    parseGameData(fetchedGameData: Object): void
    {
        // If we have received game data to display, we are no longer waiting for the HTTP request to complete.
        this.isAwaitingGameData = false;
    
        // The object fetchedGameData["ChatLog"] is only an "array-like object",
        // as is fetchedGameData["ChatLog"], and an "array-like object" is not an
        // array, so we must build an array around such an object before passing
        // it into the function to refresh a log message array.
        LogMessage.refreshListFromSource(this.chatLog, Array.from(fetchedGameData["ChatLog"]));
        LogMessage.refreshListFromSource(this.actionLog, Array.from(fetchedGameData["ActionLog"]));

        this.noncardInformationText = "Score: " + fetchedGameData["ScoreSoFar"]
         + " - Hints: " + fetchedGameData["NumberOfReadyHints"] + " / " + fetchedGameData["MaximumNumberOfHints"]
         + " - Mistakes: " + fetchedGameData["NumberOfMistakesMade"] + " / " + fetchedGameData["NumberOfMistakesIndicatingGameOver"]
         + " - Cards left in deck: " + fetchedGameData["NumberOfCardsLeftInDeck"];

         VisibleCard.refreshListOfListsFromSource(this.playedSequences, fetchedGameData["PlayedCards"]);
         VisibleCard.refreshListFromSource(this.discardPile, fetchedGameData["DiscardedCards"]);

         VisibleHand.refreshListFromSource(this.handsBeforeViewingPlayer, fetchedGameData["HandsBeforeThisPlayer"]);
         VisibleHand.refreshListFromSource(this.handsAfterViewingPlayer, fetchedGameData["HandsAfterThisPlayer"]);
    }

    hasPlayersBeforeViewingPlayer(): boolean
    {
      return this.handsBeforeViewingPlayer && (this.handsBeforeViewingPlayer.length > 0);
    }

    hasPlayersAfterViewingPlayer(): boolean
    {
      return this.handsAfterViewingPlayer && (this.handsAfterViewingPlayer.length > 0);
    }

    sendChat(): void
    {
        if (this.chatInput)
        {
            this.ilutulestikudService
            .sendChatMessage(this.gameIdentification, this.playerIdentification, this.chatInput)
            .subscribe(
                () => {},
                thrownError => this.onError.emit(thrownError),
                () => {});
            this.chatInput = null;
        }
    }
  }