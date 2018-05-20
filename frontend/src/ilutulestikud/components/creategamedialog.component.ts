import { Component } from '@angular/core';
import { Inject } from '@angular/core';
import { MatInputModule } from '@angular/material';
import { MatDialogRef } from '@angular/material';
import { MAT_DIALOG_DATA } from '@angular/material';
import { Player } from '../models/player.model'
import { Ruleset } from '../models/ruleset.model'

@Component({
    selector: 'create-game-dialog',
    templateUrl: './creategamedialog.component.html',
  })
  export class CreateGameDialogComponent
  {
    gameName: string;
    participatingPlayers: Player[];
    creatingPlayer: Player;

    // We need a copy of the available players list so that we can remove players from
    // the pool as they are added to the list of participants.
    availablePlayersCopy: Player[];
    selectedParticipant: Player;

    // We need a reference to the available rulesets list so that it can update as the
    // dialog opens.
    availableRulesetsReference: Ruleset[];
    selectedRuleset: Ruleset;

    constructor(
        public dialogReference: MatDialogRef<CreateGameDialogComponent>,
        @Inject(MAT_DIALOG_DATA) public data: any)
    {
        this.creatingPlayer = null;
        this.availablePlayersCopy = [];

        if (data)
        {
            if (data.creatingPlayer)
            {
                this.creatingPlayer = data.creatingPlayer;
            }
            
            if (data.availablePlayers)
            {
                // We need a copy of the available players list so that we can remove players from
                // the pool as they are added to the list of participants.
                this.availablePlayersCopy = data["availablePlayers"].slice();
            }
            
            if (data.availableRulesets)
            {
                // We need a reference to the available rulesets list so that it can update as the
                // dialog opens.
                this.availableRulesetsReference = data["availableRulesets"];
            }
        }

        this.selectedRuleset = null;

        this.gameName = null;
        this.participatingPlayers = this.creatingPlayer ? [this.creatingPlayer] : [];
        if (this.creatingPlayer)
        {
            this.removePlayerFromAvailablePlayerList(this.creatingPlayer);
        }
    }

    isAllowedToAddPlayer(): boolean
    {
        return (this.selectedRuleset
                && this.participatingPlayers
                && (this.participatingPlayers.length < this.selectedRuleset.MaximumNumberOfPlayers))
    }

    addParticipant(newParticipant: Player): void
    {
        this.participatingPlayers.push(newParticipant);
        this.removePlayerFromAvailablePlayerList(newParticipant);
    }

    removePlayerFromAvailablePlayerList(playerToRemove: Player)
    {
        if (!this.availablePlayersCopy)
        {
            console.log("Asked to remove player "
              + playerToRemove.ForBackend.NameForPost
              + " from non-existent available player list.");
            return;
        }

        const playerIndex = this.availablePlayersCopy.indexOf(playerToRemove);

        if (playerIndex >= 0)
        {
            this.availablePlayersCopy.splice(playerIndex, 1);
        }
    }

    isAllowedToCreateGame(): boolean
    {
        return (this.selectedRuleset
                && this.participatingPlayers
                && (this.participatingPlayers.length >= this.selectedRuleset.MinimumNumberOfPlayers))
    }

    createGame(): void
    {
        this.dialogReference.close(
            {
                "GameName": this.gameName,
                "RulesetIdentifier": this.selectedRuleset.Identifier,
                "PlayerNames": this.participatingPlayers.map(participatingPlayer => participatingPlayer.ForBackend)
            });
    }

    cancelDialog(): void
    {
        this.dialogReference.close(null);
    }
  }