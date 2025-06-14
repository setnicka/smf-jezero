package variants

import (
	"fmt"
	"html"
	"strings"
)

// CoralReef is variant for the Jezero game
type CoralReef struct{}

// NewCoralReef creates a new instance of CoralReef
func NewCoralReef() *CoralReef { return &CoralReef{} }

////////////////////////////////////////////////////////////////////////////////
//revive:disable:exported

func (cr *CoralReef) ViewTemplateName() string    { return "viewCoralReef" }
func (cr *CoralReef) TemplateStateName() string   { return "Stav moře" }
func (cr *CoralReef) TemplateStateSymbol() string { return "" }
func (cr *CoralReef) TemplateMoneyName() string   { return "Peníze" }
func (cr *CoralReef) TemplateMoneySymbol() string { return "🥒" }

func (cr *CoralReef) NopName() string        { return "Nic" }
func (cr *CoralReef) EcoName() string        { return "tradiční pěstování okurek" }
func (cr *CoralReef) HarvestName() string    { return "Průmyslové pěstování okurek" }
func (cr *CoralReef) CleaningName() string   { return "Čištění" }
func (cr *CoralReef) InspectionName() string { return "Kontrola" }
func (cr *CoralReef) EspionageName() string  { return "Špionáž" }

////////////////////////////////////////////////////////////////////////////////

func (cr *CoralReef) EcoMessage(money int, pollution int) string {
	return fmt.Sprintf("Věnovali jste se tradičnímu pěstování, získáváte %d 🥒 a zhoršili jste stav moře o %d", money, pollution)
}
func (cr *CoralReef) HarvestPenaltyMessage(penalty int) string {
	return fmt.Sprintf("Vaše průmyslové pěstování bylo odhaleno kontrolou! Nic jste nezískali a musíte místo toho zaplatit pokutu %d 🥒", penalty)
}
func (cr *CoralReef) HarvestSuccessMessage(money int, pollution int) string {
	return fmt.Sprintf("Věnovali jste se průmyslovému pěstování, získali jste za to %d 🥒 a zhoršili stav moře o %d", money, pollution)
}
func (cr *CoralReef) CleaningMessage(cleaning int) string {
	return fmt.Sprintf("Zlepšili jste čištěním stav moře o %d", cleaning)
}
func (cr *CoralReef) InspectionMessage() string {
	return "Požádali jste o kontrolu, pokud někdo v minulém kole prováděl něco špatného, tak byl potrestán"
}
func (cr *CoralReef) EspionageFailMessage() string {
	return "Špionáž nemohla být dokončena kvůli probíhající kontrole jiného týmu, nic jste nezjistili"
}
func (cr *CoralReef) EspionageSuccessMessage(teamActions map[string]string) string {
	results := []string{}
	for team, action := range teamActions {
		results = append(results, fmt.Sprintf("<li>%s: <b>%s</b></li>", html.EscapeString(team), action))
	}
	return fmt.Sprintf("Špionáž úspěšná, zjištěno:<ul>\n%s\n</ul>", strings.Join(results, "\n"))
}

func (cr *CoralReef) GlobalMessage(reduce string, increase string, change int) string {
	// increase = stav se zvedl = odsud pryč se přelilo znečištění
	return fmt.Sprintf("<b>Znečištění přes úžinu:</b> Z moře %s do moře %s se přelilo %d znečištění.", increase, reduce, change)
}
