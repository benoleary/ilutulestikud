import { Component, Input, Output, EventEmitter } from '@angular/core';
import { OnInit, OnDestroy } from '@angular/core/src/metadata/lifecycle_hooks';
import { MatDialog } from '@angular/material';
import { Observable } from 'rxjs/Rx';
import { Subscription } from 'rxjs/Subscription';
import { IlutulestikudService } from '../ilutulestikud.service';
import { LogMessage } from '../models/logmessage.model';
import { BackendIdentification } from '../models/backendidentification.model';
import { VisibleCard } from '../models/visiblecard.model';
import { VisibleHand } from '../models/visiblehand.model';
import { InferredCard } from '../models/inferredcard.model';
import { SelectHintDialogComponent } from './selecthintdialog.component'
import { HandDetailsDisplayComponent } from './handdetailsdisplay.component';


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
    gameIsFinished: boolean;
    scoreFromCardsPlayed: number;
    noncardInformationText: string;
    playedSequences: VisibleCard[][];
    discardPile: VisibleCard[];
    handsBeforeViewingPlayer: VisibleHand[];
    handOfViewingPlayer: InferredCard[];
    handsAfterViewingPlayer: VisibleHand[];
    turnButtonsDisabled: boolean;
    hintButtonsDisabled: boolean;
    colorsForHints: string[];
    indicesForHints: number[];

    constructor(public ilutulestikudService: IlutulestikudService, public materialDialog: MatDialog)
    {
        this.gameIdentification = null;
        this.playerIdentification = null;
        this.chatLog = [];
        this.chatInput = null;
        this.actionLog = [];
        this.gameIsFinished = false;
        this.scoreFromCardsPlayed = 0;
        this.noncardInformationText = null;
        this.playedSequences = [];
        this.discardPile = [];
        this.handsBeforeViewingPlayer = [];
        this.handOfViewingPlayer = [];
        this.handsAfterViewingPlayer = [];
        this.turnButtonsDisabled = true;
        this.hintButtonsDisabled = true;
        this.colorsForHints = [];
        this.indicesForHints = [];
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
          .GameAsSeenByPlayer(this.gameIdentification, this.playerIdentification)
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
        LogMessage.RefreshListFromSource(this.chatLog, Array.from(fetchedGameData["ChatLog"]));
        LogMessage.RefreshListFromSource(this.actionLog, Array.from(fetchedGameData["ActionLog"]));

        this.gameIsFinished = fetchedGameData["GameIsFinished"];
        this.scoreFromCardsPlayed = fetchedGameData["ScoreSoFar"];

        const numberOfHintsAvailable: number = fetchedGameData["NumberOfReadyHints"];
        this.noncardInformationText = "Score: " + this.scoreFromCardsPlayed
         + " - Hints: " + numberOfHintsAvailable + " / " + fetchedGameData["MaximumNumberOfHints"]
         + " - Mistakes: " + fetchedGameData["NumberOfMistakesMade"] + " / " + fetchedGameData["NumberOfMistakesIndicatingGameOver"]
         + " - Cards left in deck: " + fetchedGameData["NumberOfCardsLeftInDeck"];

        VisibleCard.RefreshListOfListsFromSource(this.playedSequences, fetchedGameData["PlayedCards"]);
        
        VisibleCard.RefreshListFromSource(this.discardPile, fetchedGameData["DiscardedCards"]);

        VisibleHand.RefreshListFromSource(this.handsBeforeViewingPlayer, fetchedGameData["HandsBeforeThisPlayer"]);

        InferredCard.RefreshListFromSource(this.handOfViewingPlayer, fetchedGameData["HandOfThisPlayer"]);

        VisibleHand.RefreshListFromSource(this.handsAfterViewingPlayer, fetchedGameData["HandsAfterThisPlayer"]);

        this.turnButtonsDisabled = !fetchedGameData["ThisPlayerCanTakeTurn"];
        this.hintButtonsDisabled = this.turnButtonsDisabled || (numberOfHintsAvailable <= 0)

        this.colorsForHints = fetchedGameData["HintColorSuits"];
        this.indicesForHints = fetchedGameData["HintSequenceIndices"];
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
            .SendChatMessage(
              this.gameIdentification,
              this.playerIdentification,
              this.chatInput)
            .subscribe(
                () => {},
                thrownError => this.onError.emit(thrownError),
                () => {});
            this.chatInput = null;
        }
    }

    discardCard(indexInHand: number): void
    {
      // We turn off the buttons to prevent messy errors
      // if the user clicks too many times too quickly.
      this.turnButtonsDisabled = true;
      this.ilutulestikudService
      .SendTakeTurnByDiscarding(
        this.gameIdentification,
        this.playerIdentification,
        indexInHand)
      .subscribe(
          () => {},
          thrownError => this.onError.emit(thrownError),
          () => {});
    }

    playCard(indexInHand: number): void
    {
      // We turn off the buttons to prevent messy errors
      // if the user clicks too many times too quickly.
      this.turnButtonsDisabled = true;
      this.ilutulestikudService
      .SendTakeTurnByAttemptingToPlay(
        this.gameIdentification,
        this.playerIdentification,
        indexInHand)
      .subscribe(
          () => {},
          thrownError => this.onError.emit(thrownError),
          () => {});
    }

    handDetailsBefore(indexOfPlayerBefore: number): void
    {
      const playerHand: VisibleHand =
       this.handsBeforeViewingPlayer[indexOfPlayerBefore];
      this.materialDialog.open(HandDetailsDisplayComponent, {
        width: '300px',
        data: {
          visibleCards: playerHand.visibleCards,
          inferredCards: playerHand.knowledgeOfCards
        }
      });
    }

    handDetailsAfter(indexOfPlayerAfter): void
    {
      const playerHand: VisibleHand =
       this.handsAfterViewingPlayer[indexOfPlayerAfter];
      this.materialDialog.open(HandDetailsDisplayComponent, {
        width: '300px',
        data: {
          visibleCards: playerHand.visibleCards,
          inferredCards: playerHand.knowledgeOfCards
        }
      });
    }

    // The button for this is only shown for players after the viewing player and
    // is only enabled when there are no players before the viewing player, so we
    // know that the index parameter is the index of the player in the list of those
    // after the viewing player.
    hintColor(indexOfPlayer): void
    {
      // We turn off the buttons to prevent messy errors
      // if the user clicks too many times too quickly.
      this.turnButtonsDisabled = true;

      let dialogRef = this.materialDialog.open(SelectHintDialogComponent, {
        width: '200px',
        data: {
          hintPossibilities: this.colorsForHints,
          hintsAreColors: true
        }
      });

      dialogRef.afterClosed().subscribe(resultFromClose =>
      {
        if (!resultFromClose)
        {
          // If the player cancelled the hint, we re-enable the buttons so that they can take their turn.
          this.turnButtonsDisabled = false;
        }
        else
        {
          this.ilutulestikudService
          .SendTakeTurnByGivingColorHint(
            this.gameIdentification,
            this.playerIdentification,
            this.handsAfterViewingPlayer[indexOfPlayer].playerName,
            resultFromClose)
            .subscribe(
                () => {},
                thrownError => this.onError.emit(thrownError),
                () => {});
        }
      });
    }

    // The button for this is only shown for players after the viewing player and
    // is only enabled when there are no players before the viewing player, so we
    // know that the index parameter is the index of the player in the list of those
    // after the viewing player.
    hintIndex(indexOfPlayer): void
    {
      // We turn off the buttons to prevent messy errors
      // if the user clicks too many times too quickly.
      this.turnButtonsDisabled = true;

      let dialogRef = this.materialDialog.open(SelectHintDialogComponent, {
        width: '200px',
        data: {
          hintPossibilities: this.indicesForHints,
          hintsAreColors: false
        }
      });

      dialogRef.afterClosed().subscribe(resultFromClose =>
      {
        if (!resultFromClose)
        {
          // If the player cancelled the hint, we re-enable the buttons so that they can take their turn.
          this.turnButtonsDisabled = false;
        }
        else
        {
          this.ilutulestikudService
          .SendTakeTurnByGivingNumberHint(
            this.gameIdentification,
            this.playerIdentification,
            this.handsAfterViewingPlayer[indexOfPlayer].playerName,
            resultFromClose)
            .subscribe(
                () => {},
                thrownError => this.onError.emit(thrownError),
                () => {});
        }
      });
    }

    leaveGame(): void
    {
      this.ilutulestikudService
      .SendLeaveGame(this.gameIdentification, this.playerIdentification)
      .subscribe(
          () => {},
          thrownError => this.onError.emit(thrownError),
          () => {});
    }
  }